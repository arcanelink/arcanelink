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
	Domain   string `mapstructure:"domain"`
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
	viper.SetDefault("server.domain", "localhost")
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
	v := viper.New()
	v.AutomaticEnv()

	// Bind environment variables explicitly
	v.BindEnv("SERVER_HOST")
	v.BindEnv("SERVER_PORT")
	v.BindEnv("SERVER_DOMAIN")
	v.BindEnv("GRPC_PORT")
	v.BindEnv("DB_HOST")
	v.BindEnv("DB_PORT")
	v.BindEnv("DB_NAME")
	v.BindEnv("DB_USER")
	v.BindEnv("DB_PASSWORD")
	v.BindEnv("DB_SSL_MODE")
	v.BindEnv("REDIS_ADDR")
	v.BindEnv("REDIS_PASSWORD")
	v.BindEnv("REDIS_DB")
	v.BindEnv("AUTH_SERVICE_ADDR")
	v.BindEnv("MESSAGE_SERVICE_ADDR")
	v.BindEnv("ROOM_SERVICE_ADDR")
	v.BindEnv("PRESENCE_SERVICE_ADDR")
	v.BindEnv("FEDERATION_SERVICE_ADDR")
	v.BindEnv("JWT_SECRET")
	v.BindEnv("JWT_EXPIRES_IN")
	v.BindEnv("HTTP_PORT")

	// Set defaults
	v.SetDefault("SERVER_HOST", "0.0.0.0")
	v.SetDefault("SERVER_PORT", 8080)
	v.SetDefault("SERVER_DOMAIN", "localhost")
	v.SetDefault("DB_SSL_MODE", "disable")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("JWT_EXPIRES_IN", 86400)

	return &Config{
		Server: ServerConfig{
			Host:     v.GetString("SERVER_HOST"),
			Port:     v.GetInt("SERVER_PORT"),
			GRPCPort: v.GetInt("GRPC_PORT"),
			Domain:   v.GetString("SERVER_DOMAIN"),
		},
		Database: DatabaseConfig{
			Host:     v.GetString("DB_HOST"),
			Port:     v.GetInt("DB_PORT"),
			Name:     v.GetString("DB_NAME"),
			User:     v.GetString("DB_USER"),
			Password: v.GetString("DB_PASSWORD"),
			SSLMode:  v.GetString("DB_SSL_MODE"),
		},
		Redis: RedisConfig{
			Addr:     v.GetString("REDIS_ADDR"),
			Password: v.GetString("REDIS_PASSWORD"),
			DB:       v.GetInt("REDIS_DB"),
		},
		Services: ServicesConfig{
			Auth:       v.GetString("AUTH_SERVICE_ADDR"),
			Message:    v.GetString("MESSAGE_SERVICE_ADDR"),
			Room:       v.GetString("ROOM_SERVICE_ADDR"),
			Presence:   v.GetString("PRESENCE_SERVICE_ADDR"),
			Federation: v.GetString("FEDERATION_SERVICE_ADDR"),
		},
		JWT: JWTConfig{
			Secret:    v.GetString("JWT_SECRET"),
			ExpiresIn: v.GetInt64("JWT_EXPIRES_IN"),
		},
	}
}
