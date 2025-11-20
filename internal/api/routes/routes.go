package routes

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/api/handlers"
	"kisanlink-erp/internal/config"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/services"
	"kisanlink-erp/internal/utils"

	pb "github.com/Kisanlink/kisanlink-ecom/proto/gen/go/collaborator/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config, aaaMiddleware *aaa.AAAMiddleware, logger interfaces.Logger) {
	// Initialize repositories
	warehouseRepo := repositories.NewWarehouseRepository(db)
	productRepo := repositories.NewProductRepository(db)
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
	productVariantRepo := repositories.NewProductVariantRepository(db)
	purchaseOrderRepo := repositories.NewPurchaseOrderRepository(db)
	grnRepo := repositories.NewGRNRepository(db)

	// Webhook repositories
	webhookRepo := repositories.NewWebhookRepository(db)

	// Initialize S3 service
	s3Service, err := services.NewS3Service(cfg)
	if err != nil {
		panic("Failed to initialize S3 service: " + err.Error())
	}

	// Initialize AAA address gRPC client (for server-to-server communication)
	// Only initialize if AAA is enabled to support local development without AAA service
	var addressClient *aaa.AddressGRPCClient
	if cfg.AAA.Enabled && cfg.AAA.GRPCAddress != "" {
		var err error
		addressClient, err = aaa.NewAddressGRPCClient(cfg.AAA.GRPCAddress)
		if err != nil {
			panic("Failed to initialize AAA address gRPC client: " + err.Error())
		}
		utils.Info("✓ AAA address gRPC client initialized successfully")
	} else {
		utils.Info("⚠️  Skipping AAA address gRPC client initialization (AAA disabled)")
		addressClient = nil
	}
	// Note: Connection will be closed when the application shuts down

	// Initialize E-commerce collaborator gRPC client
	ecommerceClient, err := newEcommerceCollaboratorClient(&cfg.Ecommerce)
	if err != nil {
		panic("Failed to initialize E-commerce collaborator gRPC client: " + err.Error())
	}

	// Initialize services
	warehouseService := services.NewWarehouseService(warehouseRepo, addressClient, logger)
	productService := services.NewProductService(productRepo, productVariantRepo, logger)
	inventoryService := services.NewInventoryService(inventoryRepo, warehouseRepo, productRepo, productVariantRepo, addressClient, logger)
	discountsService := services.NewDiscountsService(discountRepo, productRepo, warehouseRepo, logger)
	taxService := services.NewTaxService(taxRepo, logger)
	salesService := services.NewSalesService(salesRepo, productRepo, inventoryRepo, productVariantRepo, discountRepo, taxRepo, warehouseRepo, logger)
	returnsService := services.NewReturnsService(returnsRepo, salesRepo, inventoryRepo, logger)
	attachmentService := services.NewAttachmentService(attachmentRepo, s3Service, logger)
	refundPoliciesService := services.NewRefundPoliciesService(refundPoliciesRepo, logger)
	bankPaymentsService := services.NewBankPaymentsService(bankPaymentsRepo, salesRepo, returnsRepo, logger)

	// Procurement services
	ecommerceTimeout := time.Duration(cfg.Ecommerce.TimeoutSeconds) * time.Second
	if ecommerceTimeout <= 0 {
		ecommerceTimeout = 5 * time.Second
	}
	collaboratorService := services.NewCollaboratorService(
		collaboratorRepo,
		addressClient,
		s3Service,
		ecommerceClient,
		ecommerceTimeout,
		cfg.Ecommerce.AuthToken,
		logger,
	)
	productVariantService := services.NewProductVariantService(productVariantRepo, productRepo, logger)
	purchaseOrderService := services.NewPurchaseOrderService(purchaseOrderRepo, collaboratorRepo, warehouseRepo, productRepo, productVariantRepo, grnRepo, inventoryRepo, logger)
	grnService := services.NewGRNService(grnRepo, purchaseOrderRepo, warehouseRepo, productRepo, inventoryRepo, logger)

	// Webhook services
	webhookSecurityService := services.NewWebhookSecurityService(cfg.Webhook.Secret)
	webhookHistoryService := services.NewWebhookHistoryService(webhookRepo)
	ecommerceWebhookService := services.NewEcommerceWebhookService(
		purchaseOrderService,
		collaboratorRepo,
		productRepo,
		productVariantRepo,
		warehouseRepo,
		grnRepo,
		inventoryRepo,
		purchaseOrderRepo,
		addressClient,
	)

	// AAA middleware is now passed as parameter

	// Initialize handlers
	warehouseHandler := handlers.NewWarehouseHandler(warehouseService, aaaMiddleware, logger)
	productHandler := handlers.NewProductHandler(productService, aaaMiddleware, logger)
	inventoryHandler := handlers.NewInventoryHandler(inventoryService, aaaMiddleware, logger)
	discountsHandler := handlers.NewDiscountsHandler(discountsService, aaaMiddleware, logger)
	taxHandler := handlers.NewTaxHandler(taxService, aaaMiddleware, logger)
	salesHandler := handlers.NewSalesHandler(salesService, aaaMiddleware, logger)
	returnsHandler := handlers.NewReturnsHandler(returnsService, aaaMiddleware, logger)
	attachmentHandler := handlers.NewAttachmentHandler(attachmentService, aaaMiddleware, logger)
	refundPoliciesHandler := handlers.NewRefundPoliciesHandler(refundPoliciesService, aaaMiddleware, logger)
	bankPaymentsHandler := handlers.NewBankPaymentsHandler(bankPaymentsService, aaaMiddleware, logger)

	// Procurement handlers
	collaboratorHandler := handlers.NewCollaboratorHandler(collaboratorService, aaaMiddleware, logger)
	productVariantHandler := handlers.NewProductVariantHandler(productVariantService, aaaMiddleware, logger)
	purchaseOrderHandler := handlers.NewPurchaseOrderHandler(purchaseOrderService, aaaMiddleware, logger)
	grnHandler := handlers.NewGRNHandler(grnService, aaaMiddleware, logger)

	// Webhook handler (no AAA middleware - uses HMAC signature verification)
	ecommerceWebhookHandler := handlers.NewEcommerceWebhookHandler(
		ecommerceWebhookService,
		webhookSecurityService,
		webhookHistoryService,
		webhookRepo,
		aaaMiddleware,
		logger,
	)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Register all handlers
		warehouseHandler.RegisterRoutes(v1)
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
		productVariantHandler.RegisterRoutes(v1)
		purchaseOrderHandler.RegisterRoutes(v1)
		grnHandler.RegisterRoutes(v1)

		// Webhook handler
		ecommerceWebhookHandler.RegisterRoutes(v1)
	}
}

func newEcommerceCollaboratorClient(cfg *config.EcommerceConfig) (pb.CollaboratorServiceClient, error) {
	if cfg == nil || cfg.GRPCAddress == "" {
		return nil, nil
	}

	// Non-blocking gRPC client - allows server to start even if e-commerce service is unavailable
	dialOptions := []grpc.DialOption{}

	if cfg.UseTLS {
		tlsConfig := &tls.Config{}

		if cfg.CACertPath != "" {
			caPem, err := os.ReadFile(cfg.CACertPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read ecommerce CA certificate: %w", err)
			}

			certPool := x509.NewCertPool()
			if ok := certPool.AppendCertsFromPEM(caPem); !ok {
				return nil, fmt.Errorf("failed to append ecommerce CA certificate to pool")
			}

			tlsConfig.RootCAs = certPool
		}

		if cfg.ClientCertPath != "" && cfg.ClientKeyPath != "" {
			certificate, err := tls.LoadX509KeyPair(cfg.ClientCertPath, cfg.ClientKeyPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load ecommerce client certificate: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{certificate}
		}

		dialOptions = append(dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Use modern grpc.NewClient() instead of deprecated grpc.DialContext()
	// NewClient performs lazy connection - actual connection happens on first RPC call
	conn, err := grpc.NewClient(cfg.GRPCAddress, dialOptions...)
	if err != nil {
		// This only fails if client creation fails (e.g., invalid options)
		fmt.Printf("⚠️  Warning: Failed to create e-commerce gRPC client: %v\n", err)
		fmt.Println("   E-commerce webhook routes will not be available")
		return nil, nil
	}

	// Honest message - no false claims about connection status
	// Connection will be established automatically on first RPC call
	fmt.Printf("📡 E-commerce gRPC client initialized (target: %s)\n", cfg.GRPCAddress)
	fmt.Println("   Connection will be established on first RPC call")
	return pb.NewCollaboratorServiceClient(conn), nil
}
