// Package logger provides a globally configured zerolog.Logger instance for the application.
// It supports structured JSON logging for production and pretty, colorized console logging for development.
package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitGlobalLogger configures zerolog's global logger instance based on the application's configuration.
func InitGlobalLogger(cfg config.LogConfig) {
	var level zerolog.Level
	switch strings.ToUpper(cfg.Level) {
	case "DEBUG":
		level = zerolog.DebugLevel
	case "INFO":
		level = zerolog.InfoLevel
	case "WARN":
		level = zerolog.WarnLevel
	case "ERROR":
		level = zerolog.ErrorLevel
	case "FATAL":
		level = zerolog.FatalLevel
	case "PANIC":
		level = zerolog.PanicLevel
	default:
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	var writer io.Writer = os.Stdout
	if strings.ToLower(cfg.Format) == "text" {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	log.Logger = zerolog.New(writer).With().Timestamp().Caller().Logger()
}
