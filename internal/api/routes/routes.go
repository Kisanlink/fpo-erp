package routes

import (
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/api/handlers"
	"kisanlink-erp/internal/config"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config, aaaMiddleware *aaa.AAAMiddleware) {
	// Initialize repositories
	warehouseRepo := repositories.NewWarehouseRepository(db)
	productRepo := repositories.NewProductRepository(db)
	priceRepo := repositories.NewProductPriceRepository(db)
	inventoryRepo := repositories.NewInventoryRepository(db)
	salesRepo := repositories.NewSalesRepository(db)
	returnsRepo := repositories.NewReturnsRepository(db)
	attachmentRepo := repositories.NewAttachmentRepository(db)
	discountRepo := repositories.NewDiscountsRepository(db)
	taxRepo := repositories.NewTaxRepository(db)
	refundPoliciesRepo := repositories.NewRefundPoliciesRepository(db)
	bankPaymentsRepo := repositories.NewBankPaymentsRepository(db)
	webhookRepo := repositories.NewWebhookRepository(db)

	// Initialize S3 service
	s3Service, err := services.NewS3Service(cfg)
	if err != nil {
		panic("Failed to initialize S3 service: " + err.Error())
	}

	// Initialize AAA address client
	addressClient, err := aaa.NewAddressClient(cfg.AAA.ServiceURL)
	if err != nil {
		panic("Failed to initialize AAA address client: " + err.Error())
	}

	// Initialize services
	warehouseService := services.NewWarehouseService(warehouseRepo, addressClient)
	productService := services.NewProductService(productRepo, priceRepo)
	priceService := services.NewProductPriceService(priceRepo, productRepo)
	inventoryService := services.NewInventoryService(inventoryRepo, warehouseRepo, productRepo, addressClient)
	discountsService := services.NewDiscountsService(discountRepo)
	taxService := services.NewTaxService(taxRepo)
	salesService := services.NewSalesService(salesRepo, productRepo, inventoryRepo, priceRepo, discountRepo, taxRepo)
	returnsService := services.NewReturnsService(returnsRepo, salesRepo, inventoryRepo)
	attachmentService := services.NewAttachmentService(attachmentRepo, s3Service)
	refundPoliciesService := services.NewRefundPoliciesService(refundPoliciesRepo)
	bankPaymentsService := services.NewBankPaymentsService(bankPaymentsRepo, salesRepo, returnsRepo)

	// Initialize webhook services
	webhookSecurityService := services.NewWebhookSecurityService()
	webhookHistoryService := services.NewWebhookHistoryService(webhookRepo)
	ecommerceWebhookService := services.NewEcommerceWebhookService(inventoryService, inventoryRepo, salesService, productRepo, warehouseRepo, webhookHistoryService)
	webhookConfigService := services.NewWebhookConfigService(webhookRepo, webhookHistoryService)
	_ = services.NewOutboundWebhookService(webhookRepo, webhookHistoryService, webhookSecurityService, warehouseRepo, productRepo)

	// AAA middleware is now passed as parameter

	// Initialize handlers
	warehouseHandler := handlers.NewWarehouseHandler(warehouseService, aaaMiddleware)
	productHandler := handlers.NewProductHandler(productService, aaaMiddleware)
	priceHandler := handlers.NewProductPriceHandler(priceService, aaaMiddleware)
	inventoryHandler := handlers.NewInventoryHandler(inventoryService, aaaMiddleware)
	discountsHandler := handlers.NewDiscountsHandler(discountsService, aaaMiddleware)
	taxHandler := handlers.NewTaxHandler(taxService, aaaMiddleware)
	salesHandler := handlers.NewSalesHandler(salesService, aaaMiddleware)
	returnsHandler := handlers.NewReturnsHandler(returnsService, aaaMiddleware)
	attachmentHandler := handlers.NewAttachmentHandler(attachmentService, aaaMiddleware)
	refundPoliciesHandler := handlers.NewRefundPoliciesHandler(refundPoliciesService, aaaMiddleware)
	bankPaymentsHandler := handlers.NewBankPaymentsHandler(bankPaymentsService, aaaMiddleware)

	// Initialize webhook handlers
	webhookHandler := handlers.NewWebhookHandler(webhookSecurityService, ecommerceWebhookService, webhookHistoryService, webhookRepo, aaaMiddleware)
	webhookConfigHandler := handlers.NewWebhookConfigHandler(webhookConfigService, aaaMiddleware)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Register all handlers
		warehouseHandler.RegisterRoutes(v1)
		priceHandler.RegisterRoutes(v1) // Register price routes before product routes to avoid conflicts
		productHandler.RegisterRoutes(v1)
		inventoryHandler.RegisterRoutes(v1)
		discountsHandler.RegisterRoutes(v1)
		taxHandler.RegisterRoutes(v1)
		salesHandler.RegisterRoutes(v1)
		returnsHandler.RegisterRoutes(v1)
		attachmentHandler.RegisterRoutes(v1)
		refundPoliciesHandler.RegisterRoutes(v1)
		bankPaymentsHandler.RegisterRoutes(v1)

		// Register webhook routes
		webhookHandler.RegisterRoutes(v1)
		webhookConfigHandler.RegisterRoutes(v1)
	}
}
