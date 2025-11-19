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
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	AWS       AWSConfig       `mapstructure:"aws"`
	CORS      CORSConfig      `mapstructure:"cors"`
	AAA       AAAConfig       `mapstructure:"aaa"`
	Webhook   WebhookConfig   `mapstructure:"webhook"`
	Ecommerce EcommerceConfig `mapstructure:"ecommerce"`
}

type ServerConfig struct {
	HTTPPort  string `mapstructure:"http_port"`
	Mode      string `mapstructure:"mode"`
	PublicURL string `mapstructure:"public_url"`
}

type DatabaseConfig struct {
	Host        string `mapstructure:"host"`
	Port        string `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Name        string `mapstructure:"name"`
	SSLMode     string `mapstructure:"ssl_mode"`
	AutoMigrate bool   `mapstructure:"auto_migrate"` // Enable/disable database auto-migration (default: true)
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
	UsePathStyle    bool   `mapstructure:"use_path_style"` // Force path-style URLs for MinIO compatibility
}

type CORSConfig struct {
	AllowedOrigins string `mapstructure:"allowed_origins"`
	AllowedHeaders string `mapstructure:"allowed_headers"`
}

type AAAConfig struct {
	Enabled     bool   `mapstructure:"enabled"` // Enable/disable AAA authentication (default: true)
	JWTSecret   string `mapstructure:"jwt_secret"`
	APIKey      string `mapstructure:"api_key"` // API key for service-to-service authentication
	CacheTTL    int    `mapstructure:"cache_ttl"`
	GRPCAddress string `mapstructure:"grpc_address"` // gRPC address for authorization (e.g., localhost:50051)
	Timeout     int    `mapstructure:"timeout_seconds"`
}

type WebhookConfig struct {
	Secret          string `mapstructure:"secret"`            // HMAC-SHA256 shared secret for signature verification
	TimeoutSeconds  int    `mapstructure:"timeout_seconds"`   // Webhook processing timeout
	MaxPayloadBytes int64  `mapstructure:"max_payload_bytes"` // Maximum request body size
}

type EcommerceConfig struct {
	GRPCAddress    string `mapstructure:"grpc_address"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
	AuthToken      string `mapstructure:"auth_token"`
	UseTLS         bool   `mapstructure:"use_tls"`
	CACertPath     string `mapstructure:"ca_cert_path"`
	ClientCertPath string `mapstructure:"client_cert_path"`
	ClientKeyPath  string `mapstructure:"client_key_path"`
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

	// Parse AAA_ENABLED (default: true)
	aaaEnabled := true // Default to enabled for production safety
	aaaEnabledStr := os.Getenv("AAA_ENABLED")
	if aaaEnabledStr != "" {
		aaaEnabled, err = strconv.ParseBool(aaaEnabledStr)
		if err != nil {
			utils.Info("Invalid AAA_ENABLED value, defaulting to true:", aaaEnabledStr)
			aaaEnabled = true
		}
	}

	// Parse DB_AUTO_MIGRATE (default: true)
	dbAutoMigrate := true // Default to enabled for production safety
	dbAutoMigrateStr := os.Getenv("DB_AUTO_MIGRATE")
	if dbAutoMigrateStr != "" {
		dbAutoMigrate, err = strconv.ParseBool(dbAutoMigrateStr)
		if err != nil {
			utils.Info("Invalid DB_AUTO_MIGRATE value, defaulting to true:", dbAutoMigrateStr)
			dbAutoMigrate = true
		}
	}

	// Parse WEBHOOK_TIMEOUT_SECONDS (default: 30)
	webhookTimeout := 30
	webhookTimeoutStr := os.Getenv("WEBHOOK_TIMEOUT_SECONDS")
	if webhookTimeoutStr != "" {
		webhookTimeout, err = strconv.Atoi(webhookTimeoutStr)
		if err != nil {
			utils.Info("Invalid WEBHOOK_TIMEOUT_SECONDS value, defaulting to 30:", webhookTimeoutStr)
			webhookTimeout = 30
		}
	}

	// Parse WEBHOOK_MAX_PAYLOAD_BYTES (default: 10MB)
	webhookMaxPayload := int64(10485760) // 10MB default
	webhookMaxPayloadStr := os.Getenv("WEBHOOK_MAX_PAYLOAD_BYTES")
	if webhookMaxPayloadStr != "" {
		webhookMaxPayload, err := strconv.ParseInt(webhookMaxPayloadStr, 10, 64)
		if err != nil {
			utils.Info("Invalid WEBHOOK_MAX_PAYLOAD_BYTES value, defaulting to 10485760:", webhookMaxPayloadStr)
			webhookMaxPayload = 10485760
		}
		_ = webhookMaxPayload // Prevent unused variable error
	}

	// Parse Ecommerce timeout (default: 5 seconds)
	ecommerceTimeout := 5
	if ecommerceTimeoutStr := os.Getenv("ECOMMERCE_GRPC_TIMEOUT_SECONDS"); ecommerceTimeoutStr != "" {
		if parsed, err := strconv.Atoi(ecommerceTimeoutStr); err == nil && parsed > 0 {
			ecommerceTimeout = parsed
		} else if err != nil {
			utils.Info("Invalid ECOMMERCE_GRPC_TIMEOUT_SECONDS value, defaulting to 5:", ecommerceTimeoutStr)
		}
	}

	// Parse Ecommerce TLS flag
	ecommerceUseTLS := false
	if ecommerceUseTLSStr := os.Getenv("ECOMMERCE_GRPC_USE_TLS"); ecommerceUseTLSStr != "" {
		if parsed, err := strconv.ParseBool(ecommerceUseTLSStr); err == nil {
			ecommerceUseTLS = parsed
		} else {
			utils.Info("Invalid ECOMMERCE_GRPC_USE_TLS value, defaulting to false:", ecommerceUseTLSStr)
		}
	}

	// Load all environment variables using os.Getenv directly
	config := &Config{
		Server: ServerConfig{
			HTTPPort:  os.Getenv("SERVER_HTTP_PORT"),
			Mode:      os.Getenv("SERVER_MODE"),
			PublicURL: os.Getenv("SERVER_PUBLIC_URL"),
		},
		Database: DatabaseConfig{
			Host:        os.Getenv("DB_POSTGRES_HOST"),
			Port:        os.Getenv("DB_POSTGRES_PORT"),
			User:        os.Getenv("DB_POSTGRES_USER"),
			Password:    os.Getenv("DB_POSTGRES_PASSWORD"),
			Name:        os.Getenv("DB_POSTGRES_DBNAME"),
			SSLMode:     os.Getenv("DB_POSTGRES_SSLMODE"),
			AutoMigrate: dbAutoMigrate,
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
			UsePathStyle:    os.Getenv("AWS_USE_PATH_STYLE") == "true",
		},
		CORS: CORSConfig{
			AllowedOrigins: os.Getenv("CORS_ALLOWED_ORIGINS"),
			AllowedHeaders: os.Getenv("CORS_ALLOWED_HEADERS"),
		},
		AAA: AAAConfig{
			Enabled:     aaaEnabled,
			JWTSecret:   os.Getenv("AAA_JWT_SECRET"),
			APIKey:      os.Getenv("AAA_API_KEY"), // API key for service-to-service authentication
			CacheTTL:    aaaCacheTTL,
			GRPCAddress: os.Getenv("AAA_GRPC_ADDRESS"),
			Timeout:     aaaTimeout,
		},
		Webhook: WebhookConfig{
			Secret:          os.Getenv("WEBHOOK_SECRET"),
			TimeoutSeconds:  webhookTimeout,
			MaxPayloadBytes: webhookMaxPayload,
		},
		Ecommerce: EcommerceConfig{
			GRPCAddress:    os.Getenv("ECOMMERCE_GRPC_ADDRESS"),
			TimeoutSeconds: ecommerceTimeout,
			AuthToken:      os.Getenv("ECOMMERCE_GRPC_AUTH_TOKEN"),
			UseTLS:         ecommerceUseTLS,
			CACertPath:     os.Getenv("ECOMMERCE_GRPC_CA_CERT"),
			ClientCertPath: os.Getenv("ECOMMERCE_GRPC_CLIENT_CERT"),
			ClientKeyPath:  os.Getenv("ECOMMERCE_GRPC_CLIENT_KEY"),
		},
	}

	// Log warning if AAA is disabled
	if !aaaEnabled {
		utils.Info("⚠️  WARNING: AAA authentication is DISABLED - For development/testing only!")
	}

	utils.Info("Configuration loaded successfully")
	utils.Info("Database Host:", config.Database.Host)
	utils.Info("Database Port:", config.Database.Port)
	utils.Info("Database Name:", config.Database.Name)
	utils.Info("Server Mode:", config.Server.Mode)
	// Never log secrets
	if config.JWT.Secret == "" {
		utils.Info("JWT secret not set")
	}

	return config
}
