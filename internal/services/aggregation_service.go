package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"
	"kisanlink-erp/internal/utils"

	"go.uber.org/zap"
)

// AggregationService provides aggregated data for frontend optimization
type AggregationService struct {
	productRepo        *repositories.ProductRepository
	variantRepo        *repositories.ProductVariantRepository
	priceRepo          *repositories.ProductPriceRepository
	inventoryRepo      *repositories.InventoryRepository
	warehouseRepo      *repositories.WarehouseRepository
	collaboratorRepo   *repositories.CollaboratorRepository
	discountRepo       *repositories.DiscountsRepository
	taxRepo            *repositories.TaxRepository
	refundPoliciesRepo *repositories.RefundPoliciesRepository
	purchaseOrderRepo  *repositories.PurchaseOrderRepository
	grnRepo            *repositories.GRNRepository
	logger             interfaces.Logger
}

// NewAggregationService creates a new AggregationService
func NewAggregationService(
	productRepo *repositories.ProductRepository,
	variantRepo *repositories.ProductVariantRepository,
	priceRepo *repositories.ProductPriceRepository,
	inventoryRepo *repositories.InventoryRepository,
	warehouseRepo *repositories.WarehouseRepository,
	collaboratorRepo *repositories.CollaboratorRepository,
	discountRepo *repositories.DiscountsRepository,
	taxRepo *repositories.TaxRepository,
	refundPoliciesRepo *repositories.RefundPoliciesRepository,
	purchaseOrderRepo *repositories.PurchaseOrderRepository,
	grnRepo *repositories.GRNRepository,
	logger interfaces.Logger,
) *AggregationService {
	return &AggregationService{
		productRepo:        productRepo,
		variantRepo:        variantRepo,
		priceRepo:          priceRepo,
		inventoryRepo:      inventoryRepo,
		warehouseRepo:      warehouseRepo,
		collaboratorRepo:   collaboratorRepo,
		discountRepo:       discountRepo,
		taxRepo:            taxRepo,
		refundPoliciesRepo: refundPoliciesRepo,
		purchaseOrderRepo:  purchaseOrderRepo,
		grnRepo:            grnRepo,
		logger:             logger,
	}
}

// ParseIncludeOptions parses the include query parameter
func ParseIncludeOptions(includeParam string) models.IncludeOptions {
	if includeParam == "" || includeParam == "all" || includeParam == "*" {
		return models.IncludeOptions{
			Variants:      true,
			Prices:        true,
			Inventory:     true,
			Collaborators: true,
			Taxes:         true,
		}
	}

	if includeParam == "none" {
		return models.IncludeOptions{}
	}

	includes := strings.Split(includeParam, ",")
	options := models.IncludeOptions{}

	for _, include := range includes {
		switch strings.TrimSpace(strings.ToLower(include)) {
		case "variants":
			options.Variants = true
		case "prices":
			options.Prices = true
		case "inventory":
			options.Inventory = true
		case "collaborators":
			options.Collaborators = true
		case "taxes":
			options.Taxes = true
		}
	}

	return options
}

// GetProductDetail returns aggregated product details
func (s *AggregationService) GetProductDetail(productID string, req *models.ProductDetailRequest) (*models.ProductDetailResponse, error) {
	s.logger.Info("Getting aggregated product detail",
		zap.String("product_id", productID),
		zap.String("include", req.Include))

	// Parse include options
	includes := ParseIncludeOptions(req.Include)

	// Get product
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		s.logger.Error("Product not found", zap.String("product_id", productID), zap.Error(err))
		return nil, errors.NewNotFoundError("Product not found")
	}

	// Use existing ProductResponse type
	response := &models.ProductDetailResponse{
		Product: models.ProductResponse{
			ID:            product.ID,
			Name:          product.Name,
			Description:   product.Description,
			CategoryID:    product.CategoryID,
			SubcategoryID: product.SubcategoryID,
			CreatedAt:     product.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     product.UpdatedAt.Format(time.RFC3339),
		},
		Variants: []models.VariantWithAggData{},
		Metadata: models.ProductMetadata{
			ReadTimestamp: time.Now().Format(time.RFC3339),
		},
	}

	// Get variants if requested
	if includes.Variants {
		variants, err := s.variantRepo.GetByProductID(productID)
		if err != nil {
			s.logger.Warn("Failed to get variants", zap.String("product_id", productID), zap.Error(err))
		} else {
			activeVariants := 0
			totalStockValue := float64(0)

			for _, v := range variants {
				// Filter by active_only
				if req.ActiveOnly && !v.IsActive {
					continue
				}

				variantDetail := s.buildVariantDetail(&v, req, includes)

				// Filter by in_stock_only
				if req.InStockOnly && (variantDetail.StockSummary == nil || !variantDetail.StockSummary.InStock) {
					continue
				}

				response.Variants = append(response.Variants, variantDetail)

				if v.IsActive {
					activeVariants++
				}

				// Calculate stock value
				if variantDetail.StockSummary != nil {
					// Use min cost price for calculation
					totalStockValue += float64(variantDetail.StockSummary.TotalQuantity) * variantDetail.StockSummary.MinCostPrice
				}
			}

			response.Metadata.TotalVariants = len(variants)
			response.Metadata.ActiveVariants = activeVariants
			response.Metadata.TotalStockValue = totalStockValue
		}

		// Get collaborator info if variants have collaborator_ids
		if includes.Collaborators && len(response.Variants) > 0 {
			// Get first variant's collaborator if exists
			for _, v := range variants {
				if len(v.CollaboratorIDs) > 0 {
					collaboratorID := v.CollaboratorIDs[0]
					collaborator, err := s.collaboratorRepo.GetByID(collaboratorID)
					if err == nil {
						isActive := false
						if collaborator.IsActive != nil {
							isActive = *collaborator.IsActive
						}
						response.Collaborator = &models.CollaboratorResponse{
							ID:            collaborator.ID,
							CompanyName:   collaborator.CompanyName,
							ContactPerson: collaborator.ContactPerson,
							ContactNumber: collaborator.ContactNumber,
							Email:         collaborator.Email,
							GSTNumber:     collaborator.GSTNumber,
							PANNumber:     collaborator.PANNumber,
							IsActive:      isActive,
							CreatedAt:     collaborator.CreatedAt.Format(time.RFC3339),
							UpdatedAt:     collaborator.UpdatedAt.Format(time.RFC3339),
						}
					}
					break
				}
			}
		}
	}

	// Generate consistency token
	response.Metadata.ConsistencyToken = s.generateConsistencyToken(response)

	return response, nil
}

// buildVariantDetail builds detailed variant information
// Returns VariantWithAggData which embeds ProductVariantResponse with aggregation-specific computed data
func (s *AggregationService) buildVariantDetail(v *models.ProductVariant, req *models.ProductDetailRequest, includes models.IncludeOptions) models.VariantWithAggData {
	// Build the embedded ProductVariantResponse
	variantResponse := models.ProductVariantResponse{
		ID:                 v.ID,
		ProductID:          v.ProductID,
		VariantName:        v.VariantName,
		Description:        v.Description,
		Quantity:           v.Quantity,
		PackSize:           v.PackSize,
		SKU:                v.SKU,
		Barcode:            v.Barcode,
		HSNCode:            v.HSNCode,
		GSTRate:            v.GSTRate,
		CollaboratorIDs:    v.CollaboratorIDs,
		BrandName:          v.BrandName,
		DosageInstructions: v.DosageInstructions,
		UsageDetails:       v.UsageDetails,
		IsActive:           v.IsActive,
		CreatedAt:          v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          v.UpdatedAt.Format(time.RFC3339),
	}

	// Parse images from JSON
	if v.Images != nil {
		var images []string
		if err := json.Unmarshal([]byte(*v.Images), &images); err == nil {
			variantResponse.Images = images
		}
	}

	// Build the VariantWithAggData wrapper
	detail := models.VariantWithAggData{
		ProductVariantResponse: variantResponse,
	}

	// Get prices if requested - fetch from product_prices table
	if includes.Prices && s.priceRepo != nil {
		productPrices, err := s.priceRepo.GetActiveByVariantID(v.ID)
		if err == nil && len(productPrices) > 0 {
			prices := &models.VariantPrices{
				Currency:       "INR",
				HasActivePrice: false,
			}

			for _, p := range productPrices {
				priceInfo := &models.PriceInfo{
					Price:         p.Price,
					EffectiveFrom: p.EffectiveFrom.Format(time.RFC3339),
				}

				switch strings.ToUpper(p.PriceType) {
				case "MRP", "RETAIL":
					prices.RetailPrice = priceInfo
					prices.HasActivePrice = true
				case "MSP", "WHOLESALE":
					prices.WholesalePrice = priceInfo
					prices.HasActivePrice = true
				case "BULK":
					prices.BulkPrice = priceInfo
					prices.HasActivePrice = true
				}
			}

			if req.PriceType != "" && req.PriceType != "all" {
				// Filter by price type
				filteredPrices := &models.VariantPrices{
					Currency:       prices.Currency,
					HasActivePrice: false,
				}
				switch strings.ToLower(req.PriceType) {
				case "retail", "mrp":
					filteredPrices.RetailPrice = prices.RetailPrice
					filteredPrices.HasActivePrice = prices.RetailPrice != nil
				case "wholesale", "msp":
					filteredPrices.WholesalePrice = prices.WholesalePrice
					filteredPrices.HasActivePrice = prices.WholesalePrice != nil
				case "bulk":
					filteredPrices.BulkPrice = prices.BulkPrice
					filteredPrices.HasActivePrice = prices.BulkPrice != nil
				default:
					filteredPrices = prices
				}
				prices = filteredPrices
			}

			detail.PricesSummary = prices
		}
	}

	// Get inventory if requested
	if includes.Inventory {
		batches, err := s.inventoryRepo.GetBatchesByVariantOrderedByExpiry(v.ID)
		if err == nil && len(batches) > 0 {
			stockSummary := &models.StockSummary{
				TotalQuantity:     0,
				AvailableQuantity: 0,
				ReservedQuantity:  0,
				InStock:           false,
				WarehouseCount:    0,
			}

			warehouseMap := make(map[string]*models.WarehouseStock)
			var minCost, maxCost float64
			var earliestExpiry *time.Time
			first := true

			for _, batch := range batches {
				// Filter by warehouse_id if provided
				if req.WarehouseID != "" && batch.WarehouseID != req.WarehouseID {
					continue
				}

				stockSummary.TotalQuantity += batch.TotalQuantity
				stockSummary.ReservedQuantity += batch.ReservedQuantity
				stockSummary.AvailableQuantity += batch.AvailableQuantity() // Real available = total - reserved

				// Track warehouse stock
				if ws, exists := warehouseMap[batch.WarehouseID]; exists {
					ws.Quantity += batch.TotalQuantity
					ws.BatchCount++
					// Update cost price to average or use latest
				} else {
					warehouseName := batch.WarehouseID // Default to ID
					warehouse, err := s.warehouseRepo.GetByID(batch.WarehouseID)
					if err == nil {
						warehouseName = warehouse.Name
					}

					warehouseMap[batch.WarehouseID] = &models.WarehouseStock{
						WarehouseID:   batch.WarehouseID,
						WarehouseName: warehouseName,
						Quantity:      batch.TotalQuantity,
						CostPrice:     batch.CostPrice,
						ExpiryDate:    batch.ExpiryDate.Format("2006-01-02"),
						BatchCount:    1,
					}
				}

				// Track min/max cost price
				if first {
					minCost = batch.CostPrice
					maxCost = batch.CostPrice
					earliestExpiry = &batch.ExpiryDate
					first = false
				} else {
					if batch.CostPrice < minCost {
						minCost = batch.CostPrice
					}
					if batch.CostPrice > maxCost {
						maxCost = batch.CostPrice
					}
					if batch.ExpiryDate.Before(*earliestExpiry) {
						earliestExpiry = &batch.ExpiryDate
					}
				}
			}

			stockSummary.InStock = stockSummary.TotalQuantity > 0
			stockSummary.WarehouseCount = len(warehouseMap)
			stockSummary.MinCostPrice = minCost
			stockSummary.MaxCostPrice = maxCost
			if earliestExpiry != nil {
				exp := earliestExpiry.Format("2006-01-02")
				stockSummary.EarliestExpiry = &exp
			}

			detail.StockSummary = stockSummary

			// Convert warehouse map to slice
			for _, ws := range warehouseMap {
				detail.WarehouseStock = append(detail.WarehouseStock, *ws)
			}
		}
	}

	// Get tax configuration if requested - now uses variant's GSTRate (GST-only system)
	if includes.Taxes {
		// GST is split 50-50 between CGST and SGST for intra-state
		gstHalf := v.GSTRate / 2
		detail.TaxConfiguration = &models.TaxConfiguration{
			GSTRate:  v.GSTRate,
			CGSTRate: gstHalf,
			SGSTRate: gstHalf,
			HSNCode:  v.HSNCode,
		}
	}

	return detail
}

// GetVariantDetail returns aggregated variant details
func (s *AggregationService) GetVariantDetail(variantID string, include string, warehouseID string) (*models.VariantDetailResponse, error) {
	s.logger.Info("Getting aggregated variant detail",
		zap.String("variant_id", variantID),
		zap.String("include", include))

	includes := ParseIncludeOptions(include)

	variant, err := s.variantRepo.GetByID(variantID)
	if err != nil {
		return nil, errors.NewNotFoundError("Variant not found")
	}

	req := &models.ProductDetailRequest{
		Include:     include,
		WarehouseID: warehouseID,
		ActiveOnly:  false,
		InStockOnly: false,
	}

	detail := s.buildVariantDetail(variant, req, includes)

	response := &models.VariantDetailResponse{
		Variant: models.VariantDetailWithProduct{
			VariantWithAggData: detail,
		},
		Metadata: models.ResponseMetadata{
			ReadTimestamp: time.Now().Format(time.RFC3339),
		},
	}

	// Include product info if requested
	if includes.Variants {
		product, err := s.productRepo.GetByID(variant.ProductID)
		if err == nil {
			response.Variant.Product = &models.ProductResponse{
				ID:            product.ID,
				Name:          product.Name,
				Description:   product.Description,
				CategoryID:    product.CategoryID,
				SubcategoryID: product.SubcategoryID,
				CreatedAt:     product.CreatedAt.Format(time.RFC3339),
				UpdatedAt:     product.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	// Include collaborator info if requested
	if includes.Collaborators && len(variant.CollaboratorIDs) > 0 {
		collaboratorID := variant.CollaboratorIDs[0]
		collaborator, err := s.collaboratorRepo.GetByID(collaboratorID)
		if err == nil {
			isActive := false
			if collaborator.IsActive != nil {
				isActive = *collaborator.IsActive
			}
			response.Variant.Collaborator = &models.CollaboratorResponse{
				ID:            collaborator.ID,
				CompanyName:   collaborator.CompanyName,
				ContactPerson: collaborator.ContactPerson,
				ContactNumber: collaborator.ContactNumber,
				Email:         collaborator.Email,
				GSTNumber:     collaborator.GSTNumber,
				PANNumber:     collaborator.PANNumber,
				IsActive:      isActive,
				CreatedAt:     collaborator.CreatedAt.Format(time.RFC3339),
				UpdatedAt:     collaborator.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	response.Metadata.ConsistencyToken = s.generateConsistencyToken(response)

	return response, nil
}

// GetSalesContext returns aggregated sales context data
func (s *AggregationService) GetSalesContext(req *models.SalesContextRequest) (*models.SalesContextResponse, error) {
	s.logger.Info("Getting sales context",
		zap.String("warehouse_id", req.WarehouseID),
		zap.String("price_type", req.PriceType))

	response := &models.SalesContextResponse{
		AvailableInventory:     []models.InventoryWithPricing{},
		GlobalTaxConfiguration: models.GlobalTaxConfig{},
		DiscountPolicies:       []models.DiscountPolicy{},
		RefundPolicies:         []models.RefundPolicyInfo{},
		PaymentMethods:         s.getPaymentMethods(),
		Metadata:               models.SalesContextMetadata{},
	}

	// Get warehouse
	if req.WarehouseID != "" {
		warehouse, err := s.warehouseRepo.GetByID(req.WarehouseID)
		if err != nil {
			return nil, errors.NewNotFoundError("Warehouse not found")
		}

		response.Warehouse = models.WarehouseResponse{
			ID:        warehouse.ID,
			Name:      warehouse.Name,
			CreatedAt: warehouse.CreatedAt.Format(time.RFC3339),
			UpdatedAt: warehouse.UpdatedAt.Format(time.RFC3339),
		}
	}

	// Get inventory batches
	var batches []models.InventoryBatch
	var err error
	if req.WarehouseID != "" {
		// Use pagination loop to get all batches for the warehouse
		limit := utils.MaxLimit
		offset := 0
		for {
			pageBatches, total, fetchErr := s.inventoryRepo.GetBatchesByWarehouse(req.WarehouseID, limit, offset)
			if fetchErr != nil {
				err = fetchErr
				break
			}
			batches = append(batches, pageBatches...)
			if int64(offset+limit) >= total {
				break
			}
			offset += limit
		}
	} else {
		batches, err = s.inventoryRepo.GetAllBatches()
	}

	if err != nil {
		s.logger.Warn("Failed to get inventory batches", zap.Error(err))
	}

	// Build inventory with pricing
	productCount := make(map[string]bool)
	variantCount := make(map[string]bool)
	totalStockValue := float64(0)

	for _, batch := range batches {
		// Filter zero stock
		if !req.IncludeZeroStock && batch.TotalQuantity <= 0 {
			continue
		}

		// Get variant info first - we need it for tax config (GST-only system)
		variant, variantErr := s.variantRepo.GetByID(batch.VariantID)

		// Build tax config from variant's GSTRate (GST-only system)
		var taxConfig models.BatchTaxConfig
		if variantErr == nil {
			gstHalf := variant.GSTRate / 2
			taxConfig = models.BatchTaxConfig{
				GSTRate:      variant.GSTRate,
				CGSTRate:     gstHalf,
				SGSTRate:     gstHalf,
				TotalGSTRate: variant.GSTRate,
				HSNCode:      variant.HSNCode,
			}
		}

		inventoryItem := models.InventoryWithPricing{
			BatchID:          batch.ID,
			VariantID:        batch.VariantID,
			QuantityTotal:    batch.TotalQuantity,       // Total inventory in batch
			QuantityReserved: batch.ReservedQuantity,    // Reserved by pending sales
			QuantitySellable: batch.AvailableQuantity(), // Real sellable = total - reserved
			CostPrice:        batch.CostPrice,
			ExpiryDate:       batch.ExpiryDate.Format("2006-01-02"),
			TaxConfig:        taxConfig,
		}

		// Populate variant info if available
		if variantErr == nil {
			// Build ProductVariantResponse
			inventoryItem.Variant = models.ProductVariantResponse{
				ID:                 variant.ID,
				ProductID:          variant.ProductID,
				VariantName:        variant.VariantName,
				Description:        variant.Description,
				Quantity:           variant.Quantity,
				PackSize:           variant.PackSize,
				SKU:                variant.SKU,
				Barcode:            variant.Barcode,
				HSNCode:            variant.HSNCode,
				GSTRate:            variant.GSTRate,
				CollaboratorIDs:    variant.CollaboratorIDs,
				BrandName:          variant.BrandName,
				DosageInstructions: variant.DosageInstructions,
				UsageDetails:       variant.UsageDetails,
				IsActive:           variant.IsActive,
				CreatedAt:          variant.CreatedAt.Format(time.RFC3339),
				UpdatedAt:          variant.UpdatedAt.Format(time.RFC3339),
			}

			// Parse images
			if variant.Images != nil {
				var images []string
				if jsonErr := json.Unmarshal([]byte(*variant.Images), &images); jsonErr == nil {
					inventoryItem.Variant.Images = images
				}
			}

			// Get selling price from product_prices table
			if s.priceRepo != nil {
				variantPrices, priceErr := s.priceRepo.GetActiveByVariantID(variant.ID)
				if priceErr == nil && len(variantPrices) > 0 {
					priceType := strings.ToLower(req.PriceType)
					if priceType == "" {
						priceType = "retail"
					}

					for _, p := range variantPrices {
						pType := strings.ToLower(p.PriceType)
						if pType == priceType || (priceType == "retail" && pType == "mrp") ||
							(priceType == "wholesale" && pType == "msp") {
							inventoryItem.SellingPrice = &models.SellingPriceInfo{
								Price:         p.Price,
								PriceType:     p.PriceType,
								Currency:      p.Currency,
								EffectiveFrom: p.EffectiveFrom.Format(time.RFC3339),
								IsActive:      p.IsActive != nil && *p.IsActive,
							}
							break
						}
					}

					// Build alternate prices
					for _, p := range variantPrices {
						pType := strings.ToLower(p.PriceType)
						if inventoryItem.SellingPrice == nil || pType != strings.ToLower(inventoryItem.SellingPrice.PriceType) {
							inventoryItem.AlternatePrices = append(inventoryItem.AlternatePrices, models.AlternatePriceInfo{
								Price:     p.Price,
								PriceType: p.PriceType,
							})
						}
					}
				}
			}

			// Calculate margin if selling price available
			if inventoryItem.SellingPrice != nil {
				marginAmount := inventoryItem.SellingPrice.Price - batch.CostPrice
				marginPercentage := float64(0)
				if batch.CostPrice > 0 {
					marginPercentage = (marginAmount / batch.CostPrice) * 100
				}
				inventoryItem.Margin = &models.MarginInfo{
					CostPrice:        batch.CostPrice,
					SellingPrice:     inventoryItem.SellingPrice.Price,
					MarginAmount:     marginAmount,
					MarginPercentage: marginPercentage,
				}
			}

			// Get product info
			product, err := s.productRepo.GetByID(variant.ProductID)
			if err == nil {
				inventoryItem.Product = models.ProductResponse{
					ID:            product.ID,
					Name:          product.Name,
					Description:   product.Description,
					CategoryID:    product.CategoryID,
					SubcategoryID: product.SubcategoryID,
					CreatedAt:     product.CreatedAt.Format(time.RFC3339),
					UpdatedAt:     product.UpdatedAt.Format(time.RFC3339),
				}
				productCount[product.ID] = true
			}

			variantCount[variant.ID] = true
		}

		response.AvailableInventory = append(response.AvailableInventory, inventoryItem)
		totalStockValue += batch.CostPrice * float64(batch.TotalQuantity)
	}

	// Get active discounts
	discounts, err := s.discountRepo.GetDiscountsByStatus("active")
	if err == nil {
		for _, d := range discounts {
			var categories []string
			if d.ApplicableCategories != nil {
				if jsonErr := json.Unmarshal([]byte(*d.ApplicableCategories), &categories); jsonErr != nil {
					categories = []string{}
				}
			}

			response.DiscountPolicies = append(response.DiscountPolicies, models.DiscountPolicy{
				ID:                   d.ID,
				Name:                 d.Name,
				DiscountType:         string(d.DiscountType),
				DiscountValue:        d.Value,
				MinQuantity:          nil, // From discount logic
				MinAmount:            d.MinOrderValue,
				ApplicableCategories: categories,
				StartDate:            d.ValidFrom.Format(time.RFC3339),
				EndDate:              d.ValidUntil.Format(time.RFC3339),
				IsActive:             d.IsActive,
			})
		}
	}

	// GST-only tax system - taxes are on ProductVariant.GSTRate
	// No need to fetch from tax repository - GST is calculated per item based on variant's GSTRate
	response.GlobalTaxConfiguration.TaxCalculationMethod = "exclusive"
	// Note: Individual item taxes are calculated from ProductVariant.GSTRate during sales

	// Get refund policies
	refundPolicies, _, err := s.refundPoliciesRepo.GetAllRefundPolicies(utils.MaxLimit, 0)
	if err == nil {
		for _, rp := range refundPolicies {
			// Calculate refund percentage from restocking fee (100 - restocking fee)
			refundPercentage := 100.0 - rp.RestockingFee
			if refundPercentage < 0 {
				refundPercentage = 0
			}
			info := models.RefundPolicyInfo{
				ID:               rp.ID,
				Name:             rp.PolicyName,
				RefundPercentage: refundPercentage,
				ValidDays:        rp.MaxDays,
				IsActive:         true, // RefundPolicy model doesn't have IsActive field
				Description:      rp.Description,
			}
			response.RefundPolicies = append(response.RefundPolicies, info)
		}
	}

	// Set metadata
	response.Metadata = models.SalesContextMetadata{
		TotalProducts:   len(productCount),
		TotalVariants:   len(variantCount),
		TotalBatches:    len(response.AvailableInventory),
		TotalStockValue: totalStockValue,
		ReadTimestamp:   time.Now().Format(time.RFC3339),
		ExpiresAt:       time.Now().Add(5 * time.Minute),
	}

	response.Metadata.ConsistencyToken = s.generateConsistencyToken(response)

	return response, nil
}

// getPaymentMethods returns available payment methods
func (s *AggregationService) getPaymentMethods() []models.PaymentMethodInfo {
	return []models.PaymentMethodInfo{
		{ID: "PAY_CASH", Name: "Cash", Type: "cash", IsActive: true},
		{ID: "PAY_UPI", Name: "UPI", Type: "upi", IsActive: true},
		{ID: "PAY_ONLINE", Name: "Online Payment", Type: "online", IsActive: true},
	}
}

// generateConsistencyToken generates a consistency token for optimistic locking
func (s *AggregationService) generateConsistencyToken(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(jsonData)
	return "CT_" + hex.EncodeToString(hash[:8])
}

// ParsePOIncludeOptions parses the include query parameter for PO detail
func ParsePOIncludeOptions(includeParam string) map[string]bool {
	options := map[string]bool{
		"collaborator": false,
		"warehouse":    false,
		"items":        false,
		"grns":         false,
		"inventory":    false,
		"payments":     false,
		"timeline":     false,
	}

	if includeParam == "" || includeParam == "all" || includeParam == "*" {
		for k := range options {
			options[k] = true
		}
		return options
	}

	if includeParam == "none" {
		return options
	}

	includes := strings.Split(includeParam, ",")
	for _, include := range includes {
		key := strings.TrimSpace(strings.ToLower(include))
		if _, exists := options[key]; exists {
			options[key] = true
		}
	}

	return options
}

// GetPODetail returns aggregated purchase order details
func (s *AggregationService) GetPODetail(poID string, req *models.PODetailRequest) (*models.PODetailResponse, error) {
	s.logger.Info("Getting aggregated PO detail",
		zap.String("po_id", poID),
		zap.String("include", req.Include))

	// Parse include options
	includes := ParsePOIncludeOptions(req.Include)

	// Get purchase order with items
	po, err := s.purchaseOrderRepo.GetByIDWithItems(poID)
	if err != nil {
		s.logger.Error("Purchase order not found", zap.String("po_id", poID), zap.Error(err))
		return nil, errors.NewNotFoundError("Purchase order not found")
	}

	// Build basic PO info
	actualDelivery := (*string)(nil)
	if po.ActualDelivery != nil {
		ad := po.ActualDelivery.Format("2006-01-02")
		actualDelivery = &ad
	}

	pendingAmount := po.TotalAmount - po.PaidAmount
	if pendingAmount < 0 {
		pendingAmount = 0
	}

	// Get collaborator name for response
	collaboratorName := ""
	if collaborator, err := s.collaboratorRepo.GetByID(po.CollaboratorID); err == nil {
		collaboratorName = collaborator.CompanyName
	}

	// Get warehouse name for response
	warehouseName := ""
	if warehouse, err := s.warehouseRepo.GetByID(po.WarehouseID); err == nil {
		warehouseName = warehouse.Name
	}

	response := &models.PODetailResponse{
		PurchaseOrder: models.PurchaseOrderResponse{
			ID:               po.ID,
			PONumber:         po.PONumber,
			CollaboratorID:   po.CollaboratorID,
			CollaboratorName: collaboratorName,
			WarehouseID:      po.WarehouseID,
			WarehouseName:    warehouseName,
			OrderDate:        po.OrderDate.Format("2006-01-02"),
			ExpectedDelivery: po.ExpectedDelivery.Format("2006-01-02"),
			ActualDelivery:   actualDelivery,
			Status:           po.Status,
			TotalAmount:      po.TotalAmount,
			PaymentStatus:    po.PaymentStatus,
			PaidAmount:       po.PaidAmount,
			IsInterState:     po.IsInterState,
			CreatedAt:        po.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        po.UpdatedAt.Format(time.RFC3339),
		},
		Metadata: models.ResponseMetadata{
			ReadTimestamp: time.Now().Format(time.RFC3339),
		},
	}

	// Get collaborator info if requested
	if includes["collaborator"] {
		collaborator, err := s.collaboratorRepo.GetByID(po.CollaboratorID)
		if err == nil {
			isActive := false
			if collaborator.IsActive != nil {
				isActive = *collaborator.IsActive
			}
			response.Collaborator = &models.CollaboratorResponse{
				ID:            collaborator.ID,
				CompanyName:   collaborator.CompanyName,
				ContactPerson: collaborator.ContactPerson,
				ContactNumber: collaborator.ContactNumber,
				Email:         collaborator.Email,
				GSTNumber:     collaborator.GSTNumber,
				PANNumber:     collaborator.PANNumber,
				IsActive:      isActive,
				CreatedAt:     collaborator.CreatedAt.Format(time.RFC3339),
				UpdatedAt:     collaborator.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	// Get warehouse info if requested
	if includes["warehouse"] {
		warehouse, err := s.warehouseRepo.GetByID(po.WarehouseID)
		if err == nil {
			response.Warehouse = &models.WarehouseResponse{
				ID:        warehouse.ID,
				Name:      warehouse.Name,
				CreatedAt: warehouse.CreatedAt.Format(time.RFC3339),
				UpdatedAt: warehouse.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	// Get items if requested
	var totalOrderedQty int64
	var totalReceivedQty int64
	var totalPendingQty int64
	var totalOrderValue float64
	var totalReceivedValue float64
	var totalRejectedValue float64

	if includes["items"] {
		for _, item := range po.Items {
			receivedQty := int64(0)
			if item.ReceivedQuantity != nil {
				receivedQty = *item.ReceivedQuantity
			}
			pendingQty := item.Quantity - receivedQty
			if pendingQty < 0 {
				pendingQty = 0
			}

			// Build PurchaseOrderItemResponse
			itemResponse := models.PurchaseOrderItemResponse{
				ID:               item.ID,
				POID:             item.POID,
				VariantID:        item.VariantID,
				ProductName:      item.ProductName,
				ProductSKU:       item.ProductSKU,
				Quantity:         item.Quantity,
				UnitPrice:        item.UnitPrice,
				LineTotal:        item.LineTotal,
				ReceivedQuantity: item.ReceivedQuantity,
				// GST Breakdown from PO item
				BasePrice:  item.BasePrice,
				GSTRate:    item.GSTRate,
				GSTAmount:  item.GSTAmount,
				CGSTRate:   item.CGSTRate,
				CGSTAmount: item.CGSTAmount,
				SGSTRate:   item.SGSTRate,
				SGSTAmount: item.SGSTAmount,
				IGSTRate:   item.IGSTRate,
				IGSTAmount: item.IGSTAmount,
				CreatedAt:  item.CreatedAt.Format(time.RFC3339),
			}

			// Get variant info if product details not already set
			if itemResponse.ProductName == nil || itemResponse.ProductSKU == nil {
				variant, err := s.variantRepo.GetByID(item.VariantID)
				if err == nil {
					if itemResponse.ProductSKU == nil && variant.SKU != nil {
						itemResponse.ProductSKU = variant.SKU
					}
					// Get product info for name
					product, err := s.productRepo.GetByID(variant.ProductID)
					if err == nil && itemResponse.ProductName == nil {
						itemResponse.ProductName = &product.Name
					}
				}
			}

			response.Items = append(response.Items, itemResponse)

			// Update totals
			totalOrderedQty += item.Quantity
			totalReceivedQty += receivedQty
			totalPendingQty += pendingQty
			totalOrderValue += item.LineTotal
			totalReceivedValue += float64(receivedQty) * item.UnitPrice
		}
	} else {
		// Still calculate totals even if items not included
		for _, item := range po.Items {
			receivedQty := int64(0)
			if item.ReceivedQuantity != nil {
				receivedQty = *item.ReceivedQuantity
			}
			pendingQty := item.Quantity - receivedQty
			if pendingQty < 0 {
				pendingQty = 0
			}

			totalOrderedQty += item.Quantity
			totalReceivedQty += receivedQty
			totalPendingQty += pendingQty
			totalOrderValue += item.LineTotal
			totalReceivedValue += float64(receivedQty) * item.UnitPrice
		}
	}

	// Get GRNs if requested
	if includes["grns"] {
		grn, err := s.grnRepo.GetByPurchaseOrder(poID)
		if err == nil {
			grnDetail := models.GRNDetail{
				ID:            grn.ID,
				GRNNumber:     grn.GRNNumber,
				POID:          grn.POID,
				ReceivedDate:  grn.ReceivedDate.Format(time.RFC3339),
				Status:        "completed",
				QualityStatus: grn.QualityStatus,
				ReceivedBy:    grn.ReceivedBy,
				Remarks:       grn.Remarks,
			}

			// Get GRN with items for detailed information
			grnWithItems, err := s.grnRepo.GetByIDWithItems(grn.ID)
			if err == nil {
				for _, item := range grnWithItems.Items {
					grnItemDetail := models.GRNItemDetail{
						POItemID:         item.POItemID,
						VariantID:        item.VariantID,
						OrderedQuantity:  item.OrderedQuantity,
						ReceivedQuantity: item.ReceivedQuantity,
						AcceptedQuantity: item.AcceptedQuantity,
						RejectedQuantity: item.RejectedQuantity,
						UnitCost:         0, // Need to get from PO item
						TotalCost:        0,
						ExpiryDate:       item.ExpiryDate.Format("2006-01-02"),
						BatchNumber:      item.BatchNumber,
					}

					// Get unit cost from PO item
					for _, poItem := range po.Items {
						if poItem.ID == item.POItemID {
							grnItemDetail.UnitCost = poItem.UnitPrice
							grnItemDetail.TotalCost = float64(item.AcceptedQuantity) * poItem.UnitPrice
							break
						}
					}

					// Calculate rejected value
					totalRejectedValue += float64(item.RejectedQuantity) * grnItemDetail.UnitCost

					grnDetail.Items = append(grnDetail.Items, grnItemDetail)

					// Include inventory created if requested
					if includes["inventory"] && item.InventoryBatchID != nil {
						batch, err := s.inventoryRepo.GetBatchByID(*item.InventoryBatchID)
						if err == nil {
							invCreated := models.InventoryCreated{
								BatchID:     batch.ID,
								VariantID:   batch.VariantID,
								WarehouseID: batch.WarehouseID,
								Quantity:    batch.TotalQuantity,
								CostPrice:   batch.CostPrice,
								ExpiryDate:  batch.ExpiryDate.Format("2006-01-02"),
								BatchNumber: item.BatchNumber,
							}
							grnDetail.InventoryCreated = append(grnDetail.InventoryCreated, invCreated)
						}
					}
				}
			}

			response.GRNs = append(response.GRNs, grnDetail)
		}
	} else {
		// Still calculate rejected value from GRN
		rejectedAmount, err := s.grnRepo.GetTotalRejectedAmountByPO(poID)
		if err == nil {
			totalRejectedValue = rejectedAmount
		}
	}

	// Calculate summary
	completionPercentage := float64(0)
	if totalOrderedQty > 0 {
		completionPercentage = (float64(totalReceivedQty) / float64(totalOrderedQty)) * 100
	}

	fulfillmentStatus := "pending"
	if totalReceivedQty >= totalOrderedQty {
		fulfillmentStatus = "fully_received"
	} else if totalReceivedQty > 0 {
		fulfillmentStatus = "partially_received"
	}

	response.Summary = models.POSummary{
		TotalOrderValue:      totalOrderValue,
		TotalReceivedValue:   totalReceivedValue,
		TotalPendingValue:    totalOrderValue - totalReceivedValue,
		TotalRejectedValue:   totalRejectedValue,
		CompletionPercentage: completionPercentage,
		TotalItemsOrdered:    totalOrderedQty,
		TotalItemsReceived:   totalReceivedQty,
		TotalItemsPending:    totalPendingQty,
		PaymentStatus:        po.PaymentStatus,
		FulfillmentStatus:    fulfillmentStatus,
	}

	// Build timeline if requested
	if includes["timeline"] {
		response.Timeline = s.buildPOTimeline(po, response.GRNs)
	}

	// Generate consistency token
	response.Metadata.ConsistencyToken = s.generateConsistencyToken(response)

	return response, nil
}

// ParseInventoryIncludeOptions parses the include query parameter for inventory list
func ParseInventoryIncludeOptions(includeParam string) map[string]bool {
	options := map[string]bool{
		"variant":   false,
		"product":   false,
		"warehouse": false,
		"prices":    false,
		"taxes":     false,
	}

	if includeParam == "" || includeParam == "all" || includeParam == "*" {
		for k := range options {
			options[k] = true
		}
		return options
	}

	if includeParam == "none" {
		return options
	}

	includes := strings.Split(includeParam, ",")
	for _, include := range includes {
		key := strings.TrimSpace(strings.ToLower(include))
		if _, exists := options[key]; exists {
			options[key] = true
		}
	}

	return options
}

// GetInventoryList returns a paginated list of inventory batches with full context
func (s *AggregationService) GetInventoryList(req *models.InventoryListRequest) (*models.InventoryListResponse, error) {
	s.logger.Info("Getting inventory list",
		zap.String("warehouse_id", req.WarehouseID),
		zap.String("variant_id", req.VariantID),
		zap.String("product_id", req.ProductID),
		zap.Bool("in_stock_only", req.InStockOnly),
		zap.Bool("expiring_soon", req.ExpiringSoon))

	// Parse include options
	includes := ParseInventoryIncludeOptions(req.Include)

	// Set default pagination values
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// Get all batches first, then filter and paginate
	var allBatches []models.InventoryBatch
	var err error

	if req.WarehouseID != "" {
		// Use pagination loop to get all batches for the warehouse
		pLimit := utils.MaxLimit
		pOffset := 0
		for {
			pageBatches, total, fetchErr := s.inventoryRepo.GetBatchesByWarehouse(req.WarehouseID, pLimit, pOffset)
			if fetchErr != nil {
				err = fetchErr
				break
			}
			allBatches = append(allBatches, pageBatches...)
			if int64(pOffset+pLimit) >= total {
				break
			}
			pOffset += pLimit
		}
	} else if req.VariantID != "" {
		// Use pagination loop to get all batches for the variant
		pLimit := utils.MaxLimit
		pOffset := 0
		for {
			pageBatches, total, fetchErr := s.inventoryRepo.GetBatchesByVariant(req.VariantID, pLimit, pOffset)
			if fetchErr != nil {
				err = fetchErr
				break
			}
			allBatches = append(allBatches, pageBatches...)
			if int64(pOffset+pLimit) >= total {
				break
			}
			pOffset += pLimit
		}
	} else {
		allBatches, err = s.inventoryRepo.GetAllBatches()
	}

	if err != nil {
		s.logger.Error("Failed to get inventory batches", zap.Error(err))
		return nil, errors.NewInternalServerError("Failed to retrieve inventory batches")
	}

	// Apply filters
	var filteredBatches []models.InventoryBatch
	productIDs := make(map[string]bool)
	variantIDs := make(map[string]bool)
	now := time.Now()
	expiringSoonDays := 30 // Define "expiring soon" as within 30 days

	for _, batch := range allBatches {
		// Filter by variant_id if provided and not already filtered
		if req.VariantID != "" && req.WarehouseID != "" && batch.VariantID != req.VariantID {
			continue
		}

		// Filter by product_id
		if req.ProductID != "" {
			variant, err := s.variantRepo.GetByID(batch.VariantID)
			if err != nil || variant.ProductID != req.ProductID {
				continue
			}
		}

		// Filter by in_stock_only
		if req.InStockOnly && batch.AvailableQuantity() <= 0 {
			continue
		}

		// Filter by expiring_soon
		if req.ExpiringSoon {
			daysUntilExpiry := int(batch.ExpiryDate.Sub(now).Hours() / 24)
			if daysUntilExpiry > expiringSoonDays {
				continue
			}
		}

		filteredBatches = append(filteredBatches, batch)
		variantIDs[batch.VariantID] = true
	}

	// Sort batches
	sortBatches(filteredBatches, req.SortBy, req.SortOrder)

	// Calculate summary before pagination
	summary := models.InventorySummary{
		TotalBatches:       len(filteredBatches),
		TotalProducts:      0,
		TotalVariants:      len(variantIDs),
		TotalStockQuantity: 0,
		TotalStockValue:    0,
		ExpiringSoonCount:  0,
		LowStockCount:      0,
		ZeroStockCount:     0,
	}

	for _, batch := range filteredBatches {
		summary.TotalStockQuantity += batch.TotalQuantity
		summary.TotalStockValue += batch.CostPrice * float64(batch.TotalQuantity)

		daysUntilExpiry := int(batch.ExpiryDate.Sub(now).Hours() / 24)
		if daysUntilExpiry <= expiringSoonDays && daysUntilExpiry > 0 {
			summary.ExpiringSoonCount++
		}

		if batch.AvailableQuantity() <= 0 {
			summary.ZeroStockCount++
		} else if req.LowStockThreshold != nil && batch.AvailableQuantity() <= *req.LowStockThreshold {
			summary.LowStockCount++
		}
	}

	// Count unique products
	for variantID := range variantIDs {
		variant, err := s.variantRepo.GetByID(variantID)
		if err == nil {
			productIDs[variant.ProductID] = true
		}
	}
	summary.TotalProducts = len(productIDs)

	// Apply pagination
	total := len(filteredBatches)
	if offset >= total {
		filteredBatches = []models.InventoryBatch{}
	} else {
		end := offset + limit
		if end > total {
			end = total
		}
		filteredBatches = filteredBatches[offset:end]
	}

	// Build response
	var batchesWithContext []models.BatchWithContext
	for _, batch := range filteredBatches {
		batchContext := s.buildBatchContext(&batch, includes, now, expiringSoonDays)
		batchesWithContext = append(batchesWithContext, batchContext)
	}

	// Build pagination info
	hasMore := offset+limit < total
	var nextOffset *int
	if hasMore {
		next := offset + limit
		nextOffset = &next
	}

	pagination := models.InventoryPagination{
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
		NextOffset: nextOffset,
	}

	// Build filters applied map
	filtersApplied := make(map[string]interface{})
	if req.WarehouseID != "" {
		filtersApplied["warehouse_id"] = req.WarehouseID
	}
	if req.VariantID != "" {
		filtersApplied["variant_id"] = req.VariantID
	}
	if req.ProductID != "" {
		filtersApplied["product_id"] = req.ProductID
	}
	if req.InStockOnly {
		filtersApplied["in_stock_only"] = true
	}
	if req.ExpiringSoon {
		filtersApplied["expiring_soon"] = true
	}
	if req.LowStockThreshold != nil {
		filtersApplied["low_stock_threshold"] = *req.LowStockThreshold
	}
	if req.SortBy != "" {
		filtersApplied["sort_by"] = req.SortBy
	}
	if req.SortOrder != "" {
		filtersApplied["sort_order"] = req.SortOrder
	}

	response := &models.InventoryListResponse{
		Batches:    batchesWithContext,
		Pagination: pagination,
		Summary:    summary,
		Metadata: models.InventoryListMetadata{
			ReadTimestamp:  time.Now().Format(time.RFC3339),
			FiltersApplied: filtersApplied,
		},
	}

	return response, nil
}

// buildBatchContext builds a batch with full context
func (s *AggregationService) buildBatchContext(batch *models.InventoryBatch, includes map[string]bool, now time.Time, expiringSoonDays int) models.BatchWithContext {
	daysUntilExpiry := int(batch.ExpiryDate.Sub(now).Hours() / 24)

	// Calculate expiry status
	expiryStatus := "good"
	if daysUntilExpiry <= 0 {
		expiryStatus = "expired"
	} else if daysUntilExpiry <= 7 {
		expiryStatus = "critical"
	} else if daysUntilExpiry <= expiringSoonDays {
		expiryStatus = "warning"
	}

	batchContext := models.BatchWithContext{
		ID: batch.ID,
		QuantityDetails: models.QuantityDetails{
			TotalQuantity:     batch.TotalQuantity,
			AvailableQuantity: batch.AvailableQuantity(),
			ReservedQuantity:  batch.ReservedQuantity,
			SoldQuantity:      0, // Would need transaction history
			InStock:           batch.AvailableQuantity() > 0,
		},
		BatchInfo: models.BatchDetails{
			BatchNumber:     nil, // Batch number is in GRN, not in InventoryBatch
			ExpiryDate:      batch.ExpiryDate.Format("2006-01-02"),
			DaysUntilExpiry: daysUntilExpiry,
			ExpiryStatus:    expiryStatus,
		},
		Metadata: models.BatchMetadata{
			CreatedAt: batch.CreatedAt.Format(time.RFC3339),
			UpdatedAt: batch.UpdatedAt.Format(time.RFC3339),
		},
	}

	// Include warehouse if requested
	if includes["warehouse"] {
		warehouse, err := s.warehouseRepo.GetByID(batch.WarehouseID)
		if err == nil {
			batchContext.Warehouse = &models.WarehouseResponse{
				ID:        warehouse.ID,
				Name:      warehouse.Name,
				CreatedAt: warehouse.CreatedAt.Format(time.RFC3339),
				UpdatedAt: warehouse.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	// Include variant if requested
	if includes["variant"] {
		variant, err := s.variantRepo.GetByID(batch.VariantID)
		if err == nil {
			variantResponse := &models.ProductVariantResponse{
				ID:                 variant.ID,
				ProductID:          variant.ProductID,
				VariantName:        variant.VariantName,
				Description:        variant.Description,
				Quantity:           variant.Quantity,
				PackSize:           variant.PackSize,
				SKU:                variant.SKU,
				Barcode:            variant.Barcode,
				HSNCode:            variant.HSNCode,
				GSTRate:            variant.GSTRate,
				CollaboratorIDs:    variant.CollaboratorIDs,
				BrandName:          variant.BrandName,
				DosageInstructions: variant.DosageInstructions,
				UsageDetails:       variant.UsageDetails,
				IsActive:           variant.IsActive,
				CreatedAt:          variant.CreatedAt.Format(time.RFC3339),
				UpdatedAt:          variant.UpdatedAt.Format(time.RFC3339),
			}

			// Parse images
			if variant.Images != nil {
				var images []string
				if err := json.Unmarshal([]byte(*variant.Images), &images); err == nil {
					variantResponse.Images = images
				}
			}

			batchContext.Variant = variantResponse

			// Include product if requested
			if includes["product"] {
				product, err := s.productRepo.GetByID(variant.ProductID)
				if err == nil {
					batchContext.Product = &models.ProductResponse{
						ID:            product.ID,
						Name:          product.Name,
						Description:   product.Description,
						CategoryID:    product.CategoryID,
						SubcategoryID: product.SubcategoryID,
						CreatedAt:     product.CreatedAt.Format(time.RFC3339),
						UpdatedAt:     product.UpdatedAt.Format(time.RFC3339),
					}
				}
			}

			// Include prices if requested - fetch from product_prices table
			if includes["prices"] && s.priceRepo != nil {
				prices, err := s.priceRepo.GetActiveByVariantID(variant.ID)
				if err == nil && len(prices) > 0 {
					sellingPrices := &models.BatchSellingPrices{}
					for _, p := range prices {
						pType := strings.ToLower(p.PriceType)
						price := p.Price
						switch pType {
						case "retail", "mrp":
							sellingPrices.Retail = &price
						case "wholesale", "msp":
							sellingPrices.Wholesale = &price
						case "bulk":
							sellingPrices.Bulk = &price
						}
					}

					pricing := &models.BatchPricing{
						CostPrice:     batch.CostPrice,
						SellingPrices: sellingPrices,
						Currency:      "INR",
					}

					// Calculate margin if retail price available
					if sellingPrices.Retail != nil {
						marginAmount := *sellingPrices.Retail - batch.CostPrice
						marginPercentage := float64(0)
						if batch.CostPrice > 0 {
							marginPercentage = (marginAmount / batch.CostPrice) * 100
						}
						pricing.Margin = &models.BatchMargin{
							RetailMargin:           marginAmount,
							RetailMarginPercentage: marginPercentage,
						}
					}

					batchContext.Pricing = pricing
				}
			}
		}
	}

	// Include tax config if requested - now uses variant's GSTRate (GST-only system)
	if includes["taxes"] {
		// Fetch variant to get GST rate
		variant, err := s.variantRepo.GetByID(batch.VariantID)
		if err == nil {
			gstHalf := variant.GSTRate / 2
			batchContext.TaxConfig = &models.BatchTaxConfig{
				GSTRate:      variant.GSTRate,
				CGSTRate:     gstHalf,
				SGSTRate:     gstHalf,
				TotalGSTRate: variant.GSTRate,
				HSNCode:      variant.HSNCode,
			}
		}
	}

	return batchContext
}

// sortBatches sorts the batches by the specified field
func sortBatches(batches []models.InventoryBatch, sortBy, sortOrder string) {
	if sortBy == "" {
		sortBy = "expiry_date"
	}
	if sortOrder == "" {
		sortOrder = "asc"
	}

	ascending := sortOrder != "desc"

	switch sortBy {
	case "expiry_date":
		for i := 0; i < len(batches)-1; i++ {
			for j := i + 1; j < len(batches); j++ {
				if ascending {
					if batches[j].ExpiryDate.Before(batches[i].ExpiryDate) {
						batches[i], batches[j] = batches[j], batches[i]
					}
				} else {
					if batches[j].ExpiryDate.After(batches[i].ExpiryDate) {
						batches[i], batches[j] = batches[j], batches[i]
					}
				}
			}
		}
	case "quantity":
		for i := 0; i < len(batches)-1; i++ {
			for j := i + 1; j < len(batches); j++ {
				if ascending {
					if batches[j].TotalQuantity < batches[i].TotalQuantity {
						batches[i], batches[j] = batches[j], batches[i]
					}
				} else {
					if batches[j].TotalQuantity > batches[i].TotalQuantity {
						batches[i], batches[j] = batches[j], batches[i]
					}
				}
			}
		}
	case "cost_price":
		for i := 0; i < len(batches)-1; i++ {
			for j := i + 1; j < len(batches); j++ {
				if ascending {
					if batches[j].CostPrice < batches[i].CostPrice {
						batches[i], batches[j] = batches[j], batches[i]
					}
				} else {
					if batches[j].CostPrice > batches[i].CostPrice {
						batches[i], batches[j] = batches[j], batches[i]
					}
				}
			}
		}
	}
}

// buildPOTimeline builds the timeline events for a purchase order
func (s *AggregationService) buildPOTimeline(po *models.PurchaseOrder, grns []models.GRNDetail) []models.POTimelineEvent {
	var timeline []models.POTimelineEvent

	// PO created event
	timeline = append(timeline, models.POTimelineEvent{
		Timestamp:   po.CreatedAt.Format(time.RFC3339),
		Event:       "purchase_order_created",
		Description: "Purchase order " + po.PONumber + " was created",
	})

	// Status change events based on current status
	statusEvents := map[string]string{
		"confirmed":        "Purchase order was confirmed by the vendor",
		"out_for_delivery": "Order is out for delivery",
		"delivered":        "Order has been delivered",
		"verified":         "Order has been verified",
		"paid":             "Order payment completed",
	}

	statusOrder := []string{"confirmed", "out_for_delivery", "delivered", "verified", "paid"}
	currentStatusIndex := -1
	for i, s := range statusOrder {
		if s == po.Status {
			currentStatusIndex = i
			break
		}
	}

	// Add status events up to current status
	for i := 0; i <= currentStatusIndex; i++ {
		status := statusOrder[i]
		if desc, ok := statusEvents[status]; ok {
			timeline = append(timeline, models.POTimelineEvent{
				Timestamp:   po.UpdatedAt.Format(time.RFC3339), // Approximate
				Event:       "status_changed",
				Description: desc,
			})
		}
	}

	// GRN created events
	for _, grn := range grns {
		timeline = append(timeline, models.POTimelineEvent{
			Timestamp:   grn.ReceivedDate,
			Event:       "grn_created",
			Description: "Goods Receipt Note " + grn.GRNNumber + " was created",
			Actor:       &grn.ReceivedBy,
		})
	}

	// Payment events if paid
	if po.PaidAmount > 0 {
		timeline = append(timeline, models.POTimelineEvent{
			Timestamp:   po.UpdatedAt.Format(time.RFC3339),
			Event:       "payment_received",
			Description: "Payment was recorded",
		})
	}

	return timeline
}
