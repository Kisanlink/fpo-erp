package server

import (
	_ "kisanlink-erp/docs"
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/api/middleware"
	"kisanlink-erp/internal/api/routes"
	"kisanlink-erp/internal/config"
	"kisanlink-erp/internal/utils"

	"github.com/gin-gonic/gin"
	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// Server represents the HTTP API server
type Server struct {
	Router        *gin.Engine
	config        *config.Config
	db            *gorm.DB
	aaaMiddleware *aaa.AAAMiddleware
}

// NewServer creates a new API server
func NewServer(db *gorm.DB, cfg *config.Config) *Server {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create router
	router := gin.New()

	// Initialize AAA middleware
	aaaMiddleware, err := aaa.NewAAAMiddleware(cfg)
	if err != nil {
		utils.Error("Failed to initialize AAA middleware:", err)
		panic("Failed to initialize AAA middleware: " + err.Error())
	}

	// Create server instance
	server := &Server{
		Router:        router,
		config:        cfg,
		db:            db,
		aaaMiddleware: aaaMiddleware,
	}

	// Setup middleware
	server.setupMiddleware()

	// Setup routes
	server.setupRoutes()

	return server
}

// setupMiddleware configures all middleware
func (s *Server) setupMiddleware() {
	// Global middleware
	s.Router.Use(
		middleware.LoggingMiddleware(),
		middleware.RequestIDMiddleware(),
		middleware.ErrorLoggingMiddleware(),
		middleware.PerformanceMiddleware(5), // 5 second threshold
		middleware.CORSMiddleware(s.config),
		middleware.SecurityHeadersMiddleware(),
		middleware.CreateRateLimitMiddleware(100, 60), // 100 requests per minute
	)

	// Health check endpoint
	// @Summary Health check
	// @Description Check if the service is running
	// @Tags health
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Router /health [get]
	s.Router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "kisanlink-erp",
		})
	})

	// Swagger documentation route
	s.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// OpenAPI JSON spec route
	s.Router.GET("/api-docs", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.File("docs/swagger.json")
	})

	// Scalar documentation route - using go-scalar-api-reference package
	s.Router.GET("/docs", func(c *gin.Context) {
		// Build full spec URL
		scheme := "http"
		if c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := c.Request.Host
		if host == "" {
			host = "localhost:3000"
		}
		specURL := scheme + "://" + host + "/api-docs"

		// Configure Scalar options
		options := scalar.Options{
			SpecURL:            specURL,
			Theme:              scalar.ThemePurple,
			Layout:             scalar.LayoutModern,
			ShowSidebar:        true,
			HideDownloadButton: false,
			DarkMode:           false,
			WithDefaultFonts:   true,
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Kisanlink ERP API Documentation",
			},
		}

		// Generate Scalar HTML
		html, err := scalar.ApiReferenceHTML(&options)
		if err != nil {
			utils.Error("Failed to generate Scalar documentation:", err)
			c.String(500, "Failed to generate API documentation: "+err.Error())
			return
		}

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, html)
	})
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Register all routes using the routes package with the AAA middleware
	routes.RegisterRoutes(s.Router, s.db, s.config, s.aaaMiddleware)

	// 404 handler for unmatched routes
	s.Router.NoRoute(func(c *gin.Context) {
		utils.NotFoundResponse(c, "Endpoint not found")
	})
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	utils.Info("Starting HTTP server on", addr)
	return s.Router.Run(addr)
}

// Close gracefully closes all server resources
func (s *Server) Close() error {
	utils.Info("Closing server resources...")

	// Close AAA middleware connections
	if s.aaaMiddleware != nil {
		if err := s.aaaMiddleware.Close(); err != nil {
			utils.Error("Failed to close AAA middleware:", err)
			return err
		}
		utils.Info("AAA middleware connections closed successfully")
	}

	utils.Info("Server resources closed successfully")
	return nil
}
