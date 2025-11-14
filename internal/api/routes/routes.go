package routes

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/api/handlers"
	"kisanlink-erp/internal/config"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/services"

	pb "kisanlink-ecom/proto/gen/go/collaborator/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
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

	// Webhook repositories
	webhookRepo := repositories.NewWebhookRepository(db)

	// Initialize S3 service
	s3Service, err := services.NewS3Service(cfg)
	if err != nil {
		panic("Failed to initialize S3 service: " + err.Error())
	}

	// Initialize AAA address gRPC client (for server-to-server communication)
	addressClient, err := aaa.NewAddressGRPCClient(cfg.AAA.GRPCAddress)
	if err != nil {
		panic("Failed to initialize AAA address gRPC client: " + err.Error())
	}
	// Note: Connection will be closed when the application shuts down

	// Initialize E-commerce collaborator gRPC client
	ecommerceClient, err := newEcommerceCollaboratorClient(&cfg.Ecommerce)
	if err != nil {
		panic("Failed to initialize E-commerce collaborator gRPC client: " + err.Error())
	}

	// Initialize services
	warehouseService := services.NewWarehouseService(warehouseRepo, addressClient)
	productService := services.NewProductService(productRepo, priceRepo, productVariantRepo)
	priceService := services.NewProductPriceService(priceRepo, productRepo, productVariantRepo)
	inventoryService := services.NewInventoryService(inventoryRepo, warehouseRepo, productRepo, productVariantRepo, addressClient)
	discountsService := services.NewDiscountsService(discountRepo, productRepo, warehouseRepo)
	taxService := services.NewTaxService(taxRepo)
	salesService := services.NewSalesService(salesRepo, productRepo, inventoryRepo, priceRepo, discountRepo, taxRepo, warehouseRepo)
	returnsService := services.NewReturnsService(returnsRepo, salesRepo, inventoryRepo)
	attachmentService := services.NewAttachmentService(attachmentRepo, s3Service)
	refundPoliciesService := services.NewRefundPoliciesService(refundPoliciesRepo)
	bankPaymentsService := services.NewBankPaymentsService(bankPaymentsRepo, salesRepo, returnsRepo)

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
	)
	collaboratorProductService := services.NewCollaboratorProductService(collaboratorProductRepo, collaboratorRepo, productRepo, productVariantRepo)
	productVariantService := services.NewProductVariantService(productVariantRepo, productRepo)
	purchaseOrderService := services.NewPurchaseOrderService(purchaseOrderRepo, collaboratorRepo, warehouseRepo, productRepo, productVariantRepo, grnRepo, inventoryRepo)
	grnService := services.NewGRNService(grnRepo, purchaseOrderRepo, warehouseRepo, productRepo, inventoryRepo)

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

	// Webhook handler (no AAA middleware - uses HMAC signature verification)
	ecommerceWebhookHandler := handlers.NewEcommerceWebhookHandler(
		ecommerceWebhookService,
		webhookSecurityService,
		webhookHistoryService,
		webhookRepo,
		aaaMiddleware,
	)

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

		// Webhook handler
		ecommerceWebhookHandler.RegisterRoutes(v1)
	}
}

func newEcommerceCollaboratorClient(cfg *config.EcommerceConfig) (pb.CollaboratorServiceClient, error) {
	if cfg == nil || cfg.GRPCAddress == "" {
		return nil, nil
	}

	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	dialOptions := []grpc.DialOption{grpc.WithBlock()}

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

	conn, err := grpc.DialContext(ctx, cfg.GRPCAddress, dialOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial ecommerce collaborator service: %w", err)
	}

	return pb.NewCollaboratorServiceClient(conn), nil
}
