// Package config provides centralized, production-grade configuration management.
// It uses a factory pattern and relies solely on environment variables as the source of truth.
package config

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Security SecurityConfig `mapstructure:"security"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
	IdleTimeout  time.Duration `mapstructure:"idleTimeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	ConnMaxIdleTime time.Duration `mapstructure:"connMaxIdleTime"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
}

func (db *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		db.Host, db.User, db.Password, db.DBName, db.Port, db.SSLMode)
}

type SecurityConfig struct {
	TokenDuration time.Duration `mapstructure:"tokenDuration"`
	PasetoKey     string        `mapstructure:"pasetoKey"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// New creates a new Config instance by loading, binding, unmarshaling, and validating settings.
func New() (*Config, error) {
	v := viper.New()

	setDefaults(v)

	if err := bindEnvs(v, Config{}); err != nil {
		return nil, err
	}

	var loadedCfg Config
	if err := v.Unmarshal(&loadedCfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := validateCriticalConfigs(&loadedCfg); err != nil {
		return nil, err
	}

	return &loadedCfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.readTimeout", "5s")
	v.SetDefault("server.writeTimeout", "10s")
	v.SetDefault("server.idleTimeout", "120s")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.maxOpenConns", 25)
	v.SetDefault("database.maxIdleConns", 25)
	v.SetDefault("database.connMaxIdleTime", "15m")
	v.SetDefault("database.connMaxLifetime", "2h")
	v.SetDefault("security.tokenDuration", "15m")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
}

// bindEnvs uses reflection to dynamically bind environment variables to the Viper instance
// based on the struct tags in the provided configuration struct.
func bindEnvs(v *viper.Viper, iface interface{}, parts ...string) error {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	if ifv.Kind() == reflect.Ptr {
		ifv = ifv.Elem()
		ift = ift.Elem()
	}

	for i := 0; i < ift.NumField(); i++ {
		fieldv := ifv.Field(i)
		fieldt := ift.Field(i)
		name := fieldt.Tag.Get("mapstructure")
		if name == "" {
			name = strings.ToLower(fieldt.Name)
		}
		path := append(parts, name)

		if fieldv.Kind() == reflect.Struct {
			if err := bindEnvs(v, fieldv.Interface(), path...); err != nil {
				return err
			}
		} else {
			key := strings.Join(path, ".")
			envVar := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
			if err := v.BindEnv(key, envVar); err != nil {
				return fmt.Errorf("failed to bind env var %s to key %s: %w", envVar, key, err)
			}
		}
	}
	return nil
}

// validateCriticalConfigs checks for the presence of essential configuration values.
func validateCriticalConfigs(c *Config) error {
	if c.Database.User == "" {
		return fmt.Errorf("FATAL: Database user is not configured. Set DATABASE_USER environment variable")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("FATAL: Database password is not configured. Set DATABASE_PASSWORD environment variable")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("FATAL: Database name is not configured. Set DATABASE_DBNAME environment variable")
	}
	if c.Security.PasetoKey == "" {
		return fmt.Errorf("FATAL: PASETO key is not configured. Set SECURITY_PASETOKEY environment variable")
	}
	if len(c.Security.PasetoKey) != 32 {
		return fmt.Errorf("FATAL: PASETO key must be exactly 32 characters long")
	}
	return nil
}
