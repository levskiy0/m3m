package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	MongoDB MongoDBConfig `mapstructure:"mongodb"`
	JWT     JWTConfig     `mapstructure:"jwt"`
	Storage StorageConfig `mapstructure:"storage"`
	Runtime RuntimeConfig `mapstructure:"runtime"`
	Plugins PluginsConfig `mapstructure:"plugins"`
	Logging LoggingConfig `mapstructure:"logging"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	URI  string `mapstructure:"uri"`
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

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 3000)
	viper.SetDefault("server.uri", "http://127.0.0.1:3000")
	viper.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongodb.database", "m3m")
	viper.SetDefault("jwt.expiration", "168h")
	viper.SetDefault("storage.path", "./storage")
	viper.SetDefault("runtime.worker_pool_size", 10)
	viper.SetDefault("runtime.timeout", "30s")
	viper.SetDefault("plugins.path", "./plugins")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.path", "./logs")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
