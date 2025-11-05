package services

import (
	"context"
	"encoding/json"
	"fmt"
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/utils"
	"time"
)

// EcommerceWebhookService handles e-commerce webhook business logic
type EcommerceWebhookService struct {
	poService          *PurchaseOrderService
	collaboratorRepo   *repositories.CollaboratorRepository
	productRepo        *repositories.ProductRepository
	productVariantRepo *repositories.ProductVariantRepository
	warehouseRepo      *repositories.WarehouseRepository
	grnRepo            *repositories.GRNRepository
	inventoryRepo      *repositories.InventoryRepository
	poRepo             *repositories.PurchaseOrderRepository
	addressClient      *aaa.AddressHTTPClient
}

// NewEcommerceWebhookService creates a new e-commerce webhook service
func NewEcommerceWebhookService(
	poService *PurchaseOrderService,
	collaboratorRepo *repositories.CollaboratorRepository,
	productRepo *repositories.ProductRepository,
	productVariantRepo *repositories.ProductVariantRepository,
	warehouseRepo *repositories.WarehouseRepository,
	grnRepo *repositories.GRNRepository,
	inventoryRepo *repositories.InventoryRepository,
	poRepo *repositories.PurchaseOrderRepository,
	addressClient *aaa.AddressHTTPClient,
) *EcommerceWebhookService {
	return &EcommerceWebhookService{
		poService:          poService,
		collaboratorRepo:   collaboratorRepo,
		productRepo:        productRepo,
		productVariantRepo: productVariantRepo,
		warehouseRepo:      warehouseRepo,
		grnRepo:            grnRepo,
		inventoryRepo:      inventoryRepo,
		poRepo:             poRepo,
		addressClient:      addressClient,
	}
}

// ========================================
// 1. ORDER.CREATED - Most Complex
// ========================================

// ProcessOrderCreated handles order.created webhook
func (s *EcommerceWebhookService) ProcessOrderCreated(ctx context.Context, webhook *models.OrderCreatedWebhook) (string, error) {
	utils.Info("Processing order.created webhook:", webhook.Order.ExternalOrderID)

	// 1. Resolve warehouse from delivery address
	warehouse, err := s.resolveWarehouse(ctx, webhook.FPO.DeliveryAddress.AddressID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve warehouse: %w", err)
	}

	// 2. Find or create collaborator
	collaborator, err := s.findOrCreateCollaborator(ctx, &webhook.Collaborator)
	if err != nil {
		return "", fmt.Errorf("failed to find/create collaborator: %w", err)
	}

	// 3. Process order items (find or create products and variants)
	poItems := make([]*models.PurchaseOrderItem, 0, len(webhook.Items))
	var totalAmount float64

	for _, item := range webhook.Items {
		// Find or create product
		product, err := s.findOrCreateProduct(ctx, &item.Product)
		if err != nil {
			return "", fmt.Errorf("failed to find/create product %s: %w", item.Product.ExternalID, err)
		}

		// Find or create variant with smart matching
		variant, err := s.findOrCreateVariant(ctx, product.ID, &item.Variant, collaborator.ID)
		if err != nil {
			return "", fmt.Errorf("failed to find/create variant %s: %w", item.Variant.ExternalID, err)
		}

		// Create PO item (will be added after PO creation)
		lineTotal := float64(item.Quantity) * item.UnitPrice
		totalAmount += lineTotal

		productName := product.Name
		poItem := &models.PurchaseOrderItem{
			VariantID:   variant.ID,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			LineTotal:   lineTotal,
			ProductName: &productName,
			ProductSKU:  variant.SKU,
		}

		poItems = append(poItems, poItem)
	}

	// 4. Parse dates
	orderDate := time.Now()
	if webhook.Order.OrderDate != nil {
		if parsed, err := models.ParseTimestamp(*webhook.Order.OrderDate); err == nil {
			orderDate = parsed
		}
	}

	expectedDelivery, err := models.ParseTimestamp(webhook.Order.ExpectedDeliveryDate)
	if err != nil {
		return "", fmt.Errorf("invalid expected_delivery_date format: %w", err)
	}

	// 5. Generate PO Number
	poNumber, err := s.generatePONumber(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to generate PO number: %w", err)
	}

	// 6. Create Purchase Order
	po := models.NewPurchaseOrder(poNumber, collaborator.ID, warehouse.ID, orderDate, expectedDelivery)
	externalOrderID := webhook.Order.ExternalOrderID
	po.ExternalOrderID = &externalOrderID
	po.TotalAmount = totalAmount
	po.Status = "placed"
	po.PaymentStatus = "unpaid"

	if err := s.poRepo.Create(po); err != nil {
		return "", fmt.Errorf("failed to create purchase order: %w", err)
	}

	// 7. Create PO Items
	for _, poItem := range poItems {
		poItem.POID = po.ID
		// Initialize BaseModel for PO item
		poItemModel := models.NewPurchaseOrderItem(po.ID, poItem.VariantID, poItem.Quantity, poItem.UnitPrice)
		poItemModel.ProductName = poItem.ProductName
		poItemModel.ProductSKU = poItem.ProductSKU

		if err := s.poRepo.CreateItem(poItemModel); err != nil {
			return "", fmt.Errorf("failed to create PO item: %w", err)
		}
	}

	utils.Info("Successfully created purchase order:", po.PONumber, "for external order:", externalOrderID)
	return po.ID, nil
}

// ========================================
// Helper Methods
// ========================================

// resolveWarehouse finds warehouse by address_id
func (s *EcommerceWebhookService) resolveWarehouse(ctx context.Context, addressID string) (*models.Warehouse, error) {
	// Find warehouse with this address_id
	warehouses, err := s.warehouseRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to query warehouses: %w", err)
	}

	for _, wh := range warehouses {
		if wh.AddressID != nil && *wh.AddressID == addressID {
			return &wh, nil
		}
	}

	return nil, fmt.Errorf("no warehouse found with address_id: %s", addressID)
}

// findOrCreateCollaborator finds existing collaborator by external_id or creates new one
func (s *EcommerceWebhookService) findOrCreateCollaborator(ctx context.Context, webhookCollab *models.WebhookCollaborator) (*models.Collaborator, error) {
	// Try to find by external_id
	existing, err := s.collaboratorRepo.FindByExternalID(webhookCollab.ExternalID)
	if err != nil {
		return nil, fmt.Errorf("failed to search collaborator: %w", err)
	}

	if existing != nil {
		utils.Info("Found existing collaborator by external_id:", webhookCollab.ExternalID)
		return existing, nil
	}

	// Collaborator not found - create new one
	utils.Info("Creating new collaborator from webhook:", webhookCollab.ExternalID)

	// Use address_id reference from webhook (AAA is the source of truth for addresses)
	// E-commerce platform creates addresses in AAA first, then sends address_id in webhook
	addressID := webhookCollab.AddressID

	// Create new collaborator
	collaborator := models.NewCollaborator(
		webhookCollab.CompanyName,
		webhookCollab.ContactPerson,
		webhookCollab.ContactNumber,
		webhookCollab.BankAccountNo,
		webhookCollab.BankIFSC,
		addressID,
	)

	externalID := webhookCollab.ExternalID
	collaborator.ExternalID = &externalID
	collaborator.Email = webhookCollab.Email
	collaborator.GSTNumber = webhookCollab.GSTNumber
	collaborator.PANNumber = webhookCollab.PANNumber
	collaborator.BankName = webhookCollab.BankName
	collaborator.Experience = webhookCollab.Experience

	if err := s.collaboratorRepo.Create(collaborator); err != nil {
		return nil, fmt.Errorf("failed to create collaborator: %w", err)
	}

	utils.Info("Successfully created collaborator:", collaborator.ID)
	return collaborator, nil
}

// findOrCreateProduct finds existing product by external_id or creates new one
func (s *EcommerceWebhookService) findOrCreateProduct(ctx context.Context, webhookProduct *models.WebhookProduct) (*models.Product, error) {
	// Try to find by external_id
	existing, err := s.productRepo.FindByExternalID(webhookProduct.ExternalID)
	if err != nil {
		return nil, fmt.Errorf("failed to search product: %w", err)
	}

	if existing != nil {
		utils.Info("Found existing product by external_id:", webhookProduct.ExternalID)
		return existing, nil
	}

	// Product not found - create new one
	utils.Info("Creating new product from webhook:", webhookProduct.ExternalID)

	product := models.NewProduct(webhookProduct.Name, webhookProduct.Description)
	externalID := webhookProduct.ExternalID
	product.ExternalID = &externalID

	if err := s.productRepo.Create(product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	utils.Info("Successfully created product:", product.ID)
	return product, nil
}

// findOrCreateVariant finds existing variant with smart matching or creates new one
// Smart matching: external_id → SKU → create new
func (s *EcommerceWebhookService) findOrCreateVariant(
	ctx context.Context,
	productID string,
	webhookVariant *models.WebhookVariant,
	collaboratorID string,
) (*models.ProductVariant, error) {
	// Tier 1: Try to find by external_id
	existing, err := s.productVariantRepo.FindByExternalID(webhookVariant.ExternalID)
	if err != nil {
		return nil, fmt.Errorf("failed to search variant by external_id: %w", err)
	}

	if existing != nil {
		utils.Info("Found existing variant by external_id:", webhookVariant.ExternalID)
		return existing, nil
	}

	// Tier 2: Try to find by SKU
	existing, err = s.productVariantRepo.FindBySKU(webhookVariant.SKU)
	if err != nil {
		return nil, fmt.Errorf("failed to search variant by SKU: %w", err)
	}

	if existing != nil {
		utils.Info("Found existing variant by SKU:", webhookVariant.SKU, "- updating external_id")
		// Update external_id for future matching
		externalID := webhookVariant.ExternalID
		existing.ExternalID = &externalID
		if err := s.productVariantRepo.Update(existing); err != nil {
			utils.Error("Failed to update variant external_id:", err)
		}
		return existing, nil
	}

	// Tier 3: Create new variant
	utils.Info("Creating new variant from webhook:", webhookVariant.ExternalID)

	variant := models.NewProductVariant(productID, webhookVariant.Name, webhookVariant.QuantityText, webhookVariant.PackSize)
	externalID := webhookVariant.ExternalID
	variant.ExternalID = &externalID
	variant.SKU = &webhookVariant.SKU
	variant.CollaboratorID = &collaboratorID
	variant.BrandName = webhookVariant.BrandName
	variant.GSTRate = webhookVariant.GSTRate
	variant.DosageInstructions = webhookVariant.DosageInstructions
	variant.UsageDetails = webhookVariant.UsageDetails

	// Handle images JSON array
	if webhookVariant.Images != nil && len(*webhookVariant.Images) > 0 {
		imagesJSON, err := json.Marshal(*webhookVariant.Images)
		if err == nil {
			imagesStr := string(imagesJSON)
			variant.Images = &imagesStr
		}
	}

	if err := s.productVariantRepo.Create(variant); err != nil {
		return nil, fmt.Errorf("failed to create variant: %w", err)
	}

	utils.Info("Successfully created variant:", variant.ID)
	return variant, nil
}

// generatePONumber generates a unique PO number (PO-2025-0001)
func (s *EcommerceWebhookService) generatePONumber(ctx context.Context) (string, error) {
	year := time.Now().Year()
	prefix := fmt.Sprintf("PO-%d-", year)

	// Get count of POs created this year
	var count int64
	// This is a simplified version - in production, use a proper sequence or counter
	count = time.Now().UnixNano() % 10000 // Temporary solution

	poNumber := fmt.Sprintf("%s%04d", prefix, count)

	// Check if exists (race condition possible - improve in production)
	exists, err := s.poRepo.PONumberExists(poNumber)
	if err != nil {
		return "", err
	}

	if exists {
		// Retry with incremented number
		count++
		poNumber = fmt.Sprintf("%s%04d", prefix, count)
	}

	return poNumber, nil
}

// ========================================
// 2. ORDER.CONFIRMED
// ========================================

// ProcessOrderConfirmed handles order.confirmed webhook
func (s *EcommerceWebhookService) ProcessOrderConfirmed(ctx context.Context, webhook *models.OrderConfirmedWebhook) error {
	utils.Info("Processing order.confirmed webhook:", webhook.ExternalOrderID)

	// Find PO by external_order_id
	po, err := s.poRepo.FindByExternalOrderID(webhook.ExternalOrderID)
	if err != nil {
		return fmt.Errorf("failed to find purchase order: %w", err)
	}

	// Update PO status to confirmed
	po.Status = "confirmed"

	// Note: ConfirmedDate is tracked via status changes in the ERP
	// If needed, can be stored in a separate status_history table

	if err := s.poRepo.Update(po); err != nil {
		return fmt.Errorf("failed to update purchase order status: %w", err)
	}

	utils.Info("Successfully confirmed purchase order:", po.PONumber)
	return nil
}

// ========================================
// 3. ORDER.SHIPPED
// ========================================

// ProcessOrderShipped handles order.shipped webhook
func (s *EcommerceWebhookService) ProcessOrderShipped(ctx context.Context, webhook *models.OrderShippedWebhook) error {
	utils.Info("Processing order.shipped webhook:", webhook.ExternalOrderID)

	// Find PO by external_order_id
	po, err := s.poRepo.FindByExternalOrderID(webhook.ExternalOrderID)
	if err != nil {
		return fmt.Errorf("failed to find purchase order: %w", err)
	}

	// Update PO status to out_for_delivery
	po.Status = "out_for_delivery"

	// Note: Tracking information (carrier, tracking number) can be stored in a separate
	// shipment_tracking table if needed. The PO model doesn't have these fields.
	// For now, status tracking is sufficient.

	if err := s.poRepo.Update(po); err != nil {
		return fmt.Errorf("failed to update purchase order with shipping info: %w", err)
	}

	utils.Info("Successfully updated shipping status for PO:", po.PONumber)
	return nil
}

// ========================================
// 4. ORDER.DELIVERED - Second Most Complex
// ========================================

// ProcessOrderDelivered handles order.delivered webhook
// Creates GRN and inventory batches for received items
func (s *EcommerceWebhookService) ProcessOrderDelivered(ctx context.Context, webhook *models.OrderDeliveredWebhook) error {
	utils.Info("Processing order.delivered webhook:", webhook.ExternalOrderID)

	// Find PO by external_order_id
	po, err := s.poRepo.FindByExternalOrderID(webhook.ExternalOrderID)
	if err != nil {
		return fmt.Errorf("failed to find purchase order: %w", err)
	}

	// Verify PO has items loaded
	if len(po.Items) == 0 {
		// Reload PO with items if not preloaded
		po, err = s.poRepo.GetByIDWithItems(po.ID)
		if err != nil {
			return fmt.Errorf("failed to reload purchase order with items: %w", err)
		}
	}

	// Parse delivery date
	deliveryDate := time.Now()
	if webhook.DeliveryDate != nil {
		if parsed, err := models.ParseTimestamp(*webhook.DeliveryDate); err == nil {
			deliveryDate = parsed
		}
	}

	// Generate GRN number (similar to PO number generation)
	grnNumber := fmt.Sprintf("GRN-%d-%04d", time.Now().Year(), time.Now().UnixNano()%10000)

	// Determine quality status based on items
	qualityStatus := "accepted" // Will be updated based on items

	// Create GRN
	// receivedBy: webhook initiated, no specific user - use system identifier
	grn := models.NewGRN(grnNumber, po.ID, po.WarehouseID, "webhook-system", deliveryDate, qualityStatus)

	if err := s.grnRepo.Create(grn); err != nil {
		return fmt.Errorf("failed to create GRN: %w", err)
	}

	// Track overall quality status
	hasRejections := false
	hasAcceptances := false

	// Create GRN items and inventory batches
	for _, deliveryItem := range webhook.Items {
		// Find matching PO item by external_variant_id
		var matchingPOItem *models.PurchaseOrderItem
		for i := range po.Items {
			// Load variant if not already preloaded
			if po.Items[i].Variant.ID == "" {
				variant, err := s.productVariantRepo.GetByID(po.Items[i].VariantID)
				if err != nil {
					utils.Error("Failed to load variant for PO item:", err)
					continue
				}
				po.Items[i].Variant = *variant
			}

			if po.Items[i].Variant.ExternalID != nil && *po.Items[i].Variant.ExternalID == deliveryItem.ExternalVariantID {
				matchingPOItem = &po.Items[i]
				break
			}
		}

		if matchingPOItem == nil {
			utils.Error("No matching PO item found for external_variant_id:", deliveryItem.ExternalVariantID)
			continue
		}

		// Parse expiry date for GRNItem (required)
		expiryDate, err := models.ParseTimestamp(deliveryItem.ExpiryDate)
		if err != nil {
			return fmt.Errorf("failed to parse expiry date: %w", err)
		}

		// Create GRN item with correct parameters
		grnItem := models.NewGRNItem(
			grn.ID,
			matchingPOItem.ID,
			matchingPOItem.VariantID,
			matchingPOItem.Quantity,       // orderedQty
			deliveryItem.ReceivedQuantity, // receivedQty
			deliveryItem.AcceptedQuantity, // acceptedQty
			expiryDate,                    // expiryDate
		)
		grnItem.BatchNumber = &deliveryItem.BatchNumber

		// Track quality status
		if deliveryItem.RejectedQuantity > 0 {
			hasRejections = true
		}
		if deliveryItem.AcceptedQuantity > 0 {
			hasAcceptances = true
		}

		if err := s.grnRepo.CreateItem(grnItem); err != nil {
			return fmt.Errorf("failed to create GRN item: %w", err)
		}

		// Note: Inventory batch creation is handled by the GRN service when processing
		// the GRN approval workflow. The webhook just creates the GRN record.
		// The ERP user will review and approve the GRN, which then creates inventory batches.

		// Update PO item received quantity
		if err := s.poRepo.UpdateItemReceivedQuantity(matchingPOItem.ID, deliveryItem.ReceivedQuantity); err != nil {
			return fmt.Errorf("failed to update PO item received quantity: %w", err)
		}
	}

	// Update GRN quality status based on items
	qualityStatusUpdate := "accepted"
	if hasRejections && hasAcceptances {
		qualityStatusUpdate = "partial"
	} else if hasRejections {
		qualityStatusUpdate = "rejected"
	}
	if err := s.grnRepo.Update(grn.ID, map[string]interface{}{"quality_status": qualityStatusUpdate}); err != nil {
		utils.Error("Failed to update GRN quality status:", err)
	}

	// Update PO status to delivered
	po.Status = "delivered"
	// Note: ActualDelivery date can be updated if needed
	po.ActualDelivery = &deliveryDate

	if err := s.poRepo.Update(po); err != nil {
		return fmt.Errorf("failed to update purchase order to delivered: %w", err)
	}

	utils.Info("Successfully processed delivery for PO:", po.PONumber, "GRN:", grn.GRNNumber)
	return nil
}

// ========================================
// 5. ORDER.PAYMENT
// ========================================

// ProcessOrderPayment handles order.payment webhook
func (s *EcommerceWebhookService) ProcessOrderPayment(ctx context.Context, webhook *models.OrderPaymentWebhook) error {
	utils.Info("Processing order.payment webhook:", webhook.ExternalOrderID)

	// Find PO by external_order_id
	po, err := s.poRepo.FindByExternalOrderID(webhook.ExternalOrderID)
	if err != nil {
		return fmt.Errorf("failed to find purchase order: %w", err)
	}

	// Update payment information
	po.PaidAmount = webhook.PaidAmount

	// Determine payment status based on amount
	if webhook.PaidAmount >= po.TotalAmount {
		po.PaymentStatus = "paid"
	} else if webhook.PaidAmount > 0 {
		po.PaymentStatus = "partial"
	} else {
		po.PaymentStatus = "unpaid"
	}

	// Note: Payment details (method, transaction ID, remarks) can be stored in a
	// separate payments table if detailed tracking is needed. For now, the PO model
	// only tracks the amount and status.

	if err := s.poRepo.Update(po); err != nil {
		return fmt.Errorf("failed to update purchase order payment: %w", err)
	}

	utils.Info("Successfully processed payment for PO:", po.PONumber, "Amount:", webhook.PaidAmount)
	return nil
}
