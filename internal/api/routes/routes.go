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

	// Procurement repositories
	collaboratorRepo := repositories.NewCollaboratorRepository(db)
	collaboratorProductRepo := repositories.NewCollaboratorProductRepository(db)
	productVariantRepo := repositories.NewProductVariantRepository(db)
	purchaseOrderRepo := repositories.NewPurchaseOrderRepository(db)
	grnRepo := repositories.NewGRNRepository(db)

	// Initialize S3 service
	s3Service, err := services.NewS3Service(cfg)
	if err != nil {
		panic("Failed to initialize S3 service: " + err.Error())
	}

	// Initialize AAA address HTTP client (mock if AAA disabled)
	var addressClient *aaa.AddressHTTPClient
	if cfg.AAA.Enabled {
		addressClient = aaa.NewAddressHTTPClient(cfg.AAA.BaseURL)
	} else {
		addressClient = aaa.NewMockAddressHTTPClient()
		// Note: Mock client returns mock addresses without making HTTP calls
	}

	// Initialize services
	warehouseService := services.NewWarehouseService(warehouseRepo, addressClient)
	productService := services.NewProductService(productRepo, priceRepo)
	priceService := services.NewProductPriceService(priceRepo, productRepo)
	inventoryService := services.NewInventoryService(inventoryRepo, warehouseRepo, productRepo, addressClient)
	discountsService := services.NewDiscountsService(discountRepo, productRepo, warehouseRepo)
	taxService := services.NewTaxService(taxRepo)
	salesService := services.NewSalesService(salesRepo, productRepo, inventoryRepo, priceRepo, discountRepo, taxRepo, warehouseRepo)
	returnsService := services.NewReturnsService(returnsRepo, salesRepo, inventoryRepo)
	attachmentService := services.NewAttachmentService(attachmentRepo, s3Service)
	refundPoliciesService := services.NewRefundPoliciesService(refundPoliciesRepo)
	bankPaymentsService := services.NewBankPaymentsService(bankPaymentsRepo, salesRepo, returnsRepo)

	// Procurement services
	collaboratorService := services.NewCollaboratorService(collaboratorRepo, addressClient, s3Service)
	collaboratorProductService := services.NewCollaboratorProductService(collaboratorProductRepo, collaboratorRepo, productRepo)
	productVariantService := services.NewProductVariantService(productVariantRepo, productRepo)
	purchaseOrderService := services.NewPurchaseOrderService(purchaseOrderRepo, collaboratorRepo, warehouseRepo, productRepo, grnRepo, inventoryRepo)
	grnService := services.NewGRNService(grnRepo, purchaseOrderRepo, warehouseRepo, productRepo, inventoryRepo)

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

	// Procurement handlers
	collaboratorHandler := handlers.NewCollaboratorHandler(collaboratorService, aaaMiddleware)
	collaboratorProductHandler := handlers.NewCollaboratorProductHandler(collaboratorProductService, aaaMiddleware)
	productVariantHandler := handlers.NewProductVariantHandler(productVariantService, aaaMiddleware)
	purchaseOrderHandler := handlers.NewPurchaseOrderHandler(purchaseOrderService, aaaMiddleware)
	grnHandler := handlers.NewGRNHandler(grnService, aaaMiddleware)

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

		// Procurement handlers
		collaboratorHandler.RegisterRoutes(v1)
		collaboratorProductHandler.RegisterRoutes(v1)
		productVariantHandler.RegisterRoutes(v1)
		purchaseOrderHandler.RegisterRoutes(v1)
		grnHandler.RegisterRoutes(v1)
	}
}
