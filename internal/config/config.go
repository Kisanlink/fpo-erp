package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"kisanlink-erp/internal/utils"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	AWS      AWSConfig      `mapstructure:"aws"`
	CORS     CORSConfig     `mapstructure:"cors"`
	AAA      AAAConfig      `mapstructure:"aaa"`
}

type ServerConfig struct {
	HTTPPort string `mapstructure:"http_port"`
	Mode     string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expiry int    `mapstructure:"expiry_hours"`
}

type AWSConfig struct {
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	S3Bucket        string `mapstructure:"s3_bucket"`
}

type CORSConfig struct {
	AllowedOrigins string `mapstructure:"allowed_origins"`
	AllowedHeaders string `mapstructure:"allowed_headers"`
}

type AAAConfig struct {
	JWTSecret   string `mapstructure:"jwt_secret"`
	CacheTTL    int    `mapstructure:"cache_ttl"`
	ServiceURL  string `mapstructure:"service_url"`
	GRPCAddress string `mapstructure:"grpc_address"`
	Timeout     int    `mapstructure:"timeout_seconds"`
}

// GetAllowedOrigins returns the allowed origins as a slice
func (c *CORSConfig) GetAllowedOrigins() []string {
	if c.AllowedOrigins == "" {
		return []string{}
	}
	return strings.Split(c.AllowedOrigins, ",")
}

// GetAllowedHeaders returns the allowed headers as a slice
func (c *CORSConfig) GetAllowedHeaders() []string {
	if c.AllowedHeaders == "" {
		return []string{}
	}
	return strings.Split(c.AllowedHeaders, ",")
}

func Load() *Config {
	// Load .env file
	godotenv.Load()

	utils.Info("Loading configuration from environment variables...")

	// Convert JWT_EXPIRY_HOURS to int
	jwtExpiryStr := os.Getenv("JWT_EXPIRY_HOURS")
	jwtExpiry, err := strconv.Atoi(jwtExpiryStr)
	if err != nil {
		panic(fmt.Sprintf("CRITICAL: Environment variable JWT_EXPIRY_HOURS must be a valid integer, got: %s", jwtExpiryStr))
	}

	// Convert AAA_CACHE_TTL to int
	aaaCacheTTLStr := os.Getenv("AAA_CACHE_TTL")
	aaaCacheTTL, err := strconv.Atoi(aaaCacheTTLStr)
	if err != nil {
		panic(fmt.Sprintf("CRITICAL: Environment variable AAA_CACHE_TTL must be a valid integer, got: %s", aaaCacheTTLStr))
	}

	// Convert AAA_TIMEOUT to int
	aaaTimeoutStr := os.Getenv("AAA_TIMEOUT_SECONDS")
	aaaTimeout, err := strconv.Atoi(aaaTimeoutStr)
	if err != nil {
		// Default to 30 seconds if not specified
		aaaTimeout = 30
	}

	// Load all environment variables using os.Getenv directly
	config := &Config{
		Server: ServerConfig{
			HTTPPort: os.Getenv("SERVER_HTTP_PORT"),
			Mode:     os.Getenv("SERVER_MODE"),
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_POSTGRES_HOST"),
			Port:     os.Getenv("DB_POSTGRES_PORT"),
			User:     os.Getenv("DB_POSTGRES_USER"),
			Password: os.Getenv("DB_POSTGRES_PASSWORD"),
			Name:     os.Getenv("DB_POSTGRES_DBNAME"),
			SSLMode:  os.Getenv("DB_POSTGRES_SSLMODE"),
		},
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
			Expiry: jwtExpiry,
		},
		AWS: AWSConfig{
			Region:          os.Getenv("AWS_REGION"),
			AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			S3Bucket:        os.Getenv("AWS_S3_BUCKET"),
		},
		CORS: CORSConfig{
			AllowedOrigins: os.Getenv("CORS_ALLOWED_ORIGINS"),
			AllowedHeaders: os.Getenv("CORS_ALLOWED_HEADERS"),
		},
		AAA: AAAConfig{
			JWTSecret:   os.Getenv("AAA_JWT_SECRET"),
			CacheTTL:    aaaCacheTTL,
			ServiceURL:  os.Getenv("AAA_SERVICE_URL"),
			GRPCAddress: os.Getenv("AAA_GRPC_ADDRESS"),
			Timeout:     aaaTimeout,
		},
	}

	utils.Info("Configuration loaded successfully")
	utils.Info("Database Host:", config.Database.Host)
	utils.Info("Database Port:", config.Database.Port)
	utils.Info("Database Name:", config.Database.Name)
	utils.Info("Server Mode:", config.Server.Mode)
	utils.Info("JWT Secret:", config.JWT.Secret)

	return config
}
