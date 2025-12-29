// Package main is the entry point for the Clinic OS Core Platform API.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/config"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/database"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/logger"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/security"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam"
	iamHttp "github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/delivery/http"
	iamStore "github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/store"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/router"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	// 1. Load environment variables from .env file for local development.
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("No .env file found, relying on system environment variables.")
	}

	// 2. Initialize configuration.
	appConfig, err := config.New()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// 3. Initialize platform services (logger, database).
	logger.InitGlobalLogger(appConfig.Log)
	log.Info().Msg("Logger initialized.")

	dbProvider, err := database.NewProvider(appConfig.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize database provider")
	}
	defer dbProvider.Close()
	log.Info().Msg("Database provider initialized.")

	// 3. Initialize security services
	tokenManager, err := security.NewPasetoManager(appConfig.Security)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create token manager")
	}
	log.Info().Msg("Security provider initialized.")

	iamRepo := iamStore.NewPgxRepository(dbProvider.Pool)
	iamSvc := iam.NewService(iamRepo, tokenManager, appConfig)
	iamHandler := iamHttp.NewHandler(iamSvc)
	log.Info().Msg("IAM module initialized.")

	// 4. Setup router with injected dependencies.
	engine := router.New(dbProvider, tokenManager, iamHandler)
	log.Info().Msg("Router initialized.")

	// 5. Create and configure the HTTP server.
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", appConfig.Server.Port),
		Handler:      engine,
		ReadTimeout:  appConfig.Server.ReadTimeout,
		WriteTimeout: appConfig.Server.WriteTimeout,
		IdleTimeout:  appConfig.Server.IdleTimeout,
	}

	// 6. Start the server and listen for shutdown signals.
	serverErrChan := make(chan error, 1)
	go func() {
		log.Info().Str("address", httpServer.Addr).Msg("Starting HTTP server")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
		close(serverErrChan)
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrChan:
		if err != nil {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	case sig := <-shutdownChan:
		log.Info().Str("signal", sig.String()).Msg("Shutdown signal received, starting graceful shutdown...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("HTTP server graceful shutdown failed")
		} else {
			log.Info().Msg("HTTP server shut down gracefully.")
		}
	}

	log.Info().Msg("Application has shut down.")
}
