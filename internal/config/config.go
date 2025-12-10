package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	MongoDB  MongoDBConfig  `mapstructure:"mongodb"`
	SQLite   SQLiteConfig   `mapstructure:"sqlite"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Runtime  RuntimeConfig  `mapstructure:"runtime"`
	Plugins  PluginsConfig  `mapstructure:"plugins"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	URI  string `mapstructure:"uri"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"` // "mongodb" or "sqlite"
}

type SQLiteConfig struct {
	Path     string `mapstructure:"path"`     // Path to SQLite data directory
	Database string `mapstructure:"database"` // Database name
}

type MongoDBConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Expiration time.Duration `mapstructure:"expiration"`
}

type StorageConfig struct {
	Path string `mapstructure:"path"`
}

type RuntimeConfig struct {
	WorkerPoolSize int           `mapstructure:"worker_pool_size"`
	Timeout        time.Duration `mapstructure:"timeout"`
}

type PluginsConfig struct {
	Path   string                            `mapstructure:"path"`
	Config map[string]map[string]interface{} `mapstructure:"config"` // Plugin-specific configs: plugins.config.$pluginName
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"path"`
}

// generateJWTSecret generates a random 32-byte hex string for JWT signing
func generateJWTSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a less secure but still random string
		return fmt.Sprintf("m3m-secret-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// createDefaultConfig creates a default config.yaml file
func createDefaultConfig(path string) error {
	jwtSecret := generateJWTSecret()

	content := fmt.Sprintf(`server:
  host: "0.0.0.0"
  port: 3000
  uri: "http://127.0.0.1:3000"

database:
  driver: "sqlite"  # "mongodb" or "sqlite"

mongodb:
  uri: "mongodb://localhost:27017"
  database: "m3m"

sqlite:
  path: "./data"
  database: "m3m"

jwt:
  secret: "%s"
  expiration: 168h

storage:
  path: "./storage"

runtime:
  worker_pool_size: 10
  timeout: 30s

plugins:
  path: "./plugins"

logging:
  level: "info"
  path: "./logs"
`, jwtSecret)

	return os.WriteFile(path, []byte(content), 0644)
}

func Load(path string) (*Config, error) {
	// Check if config file exists, create default if not
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Config file not found, creating default: %s\n", path)
		if err := createDefaultConfig(path); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		fmt.Println("Default config created with SQLite database (no external dependencies)")
	}

	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix("M3M")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 3000)
	viper.SetDefault("server.uri", "http://127.0.0.1:3000")
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongodb.database", "m3m")
	viper.SetDefault("sqlite.path", "./data")
	viper.SetDefault("sqlite.database", "m3m")
	viper.SetDefault("jwt.expiration", "168h")
	viper.SetDefault("storage.path", "./storage")
	viper.SetDefault("runtime.worker_pool_size", 10)
	viper.SetDefault("runtime.timeout", "30s")
	viper.SetDefault("plugins.path", "./plugins")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.path", "./logs")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			if path != "" {
				return nil, err
			}
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
