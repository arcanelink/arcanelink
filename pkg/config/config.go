package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Services ServicesConfig `mapstructure:"services"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	GRPCPort int    `mapstructure:"grpc_port"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// ServicesConfig holds service addresses
type ServicesConfig struct {
	Auth       string `mapstructure:"auth"`
	Message    string `mapstructure:"message"`
	Room       string `mapstructure:"room"`
	Presence   string `mapstructure:"presence"`
	Federation string `mapstructure:"federation"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpiresIn  int64  `mapstructure:"expires_in"` // seconds
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("jwt.expires_in", 86400) // 24 hours

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, use environment variables
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadFromEnv loads configuration from environment variables only
func LoadFromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			Host:     viper.GetString("SERVER_HOST"),
			Port:     viper.GetInt("SERVER_PORT"),
			GRPCPort: viper.GetInt("GRPC_PORT"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetInt("DB_PORT"),
			Name:     viper.GetString("DB_NAME"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			SSLMode:  viper.GetString("DB_SSL_MODE"),
		},
		Redis: RedisConfig{
			Addr:     viper.GetString("REDIS_ADDR"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		Services: ServicesConfig{
			Auth:       viper.GetString("AUTH_SERVICE_ADDR"),
			Message:    viper.GetString("MESSAGE_SERVICE_ADDR"),
			Room:       viper.GetString("ROOM_SERVICE_ADDR"),
			Presence:   viper.GetString("PRESENCE_SERVICE_ADDR"),
			Federation: viper.GetString("FEDERATION_SERVICE_ADDR"),
		},
		JWT: JWTConfig{
			Secret:    viper.GetString("JWT_SECRET"),
			ExpiresIn: viper.GetInt64("JWT_EXPIRES_IN"),
		},
	}
}
