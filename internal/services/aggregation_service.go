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

	"go.uber.org/zap"
)

// AggregationService provides aggregated data for frontend optimization
type AggregationService struct {
	productRepo         *repositories.ProductRepository
	variantRepo         *repositories.ProductVariantRepository
	inventoryRepo       *repositories.InventoryRepository
	warehouseRepo       *repositories.WarehouseRepository
	collaboratorRepo    *repositories.CollaboratorRepository
	discountRepo        *repositories.DiscountsRepository
	taxRepo             *repositories.TaxRepository
	refundPoliciesRepo  *repositories.RefundPoliciesRepository
	logger              interfaces.Logger
}

// NewAggregationService creates a new AggregationService
func NewAggregationService(
	productRepo *repositories.ProductRepository,
	variantRepo *repositories.ProductVariantRepository,
	inventoryRepo *repositories.InventoryRepository,
	warehouseRepo *repositories.WarehouseRepository,
	collaboratorRepo *repositories.CollaboratorRepository,
	discountRepo *repositories.DiscountsRepository,
	taxRepo *repositories.TaxRepository,
	refundPoliciesRepo *repositories.RefundPoliciesRepository,
	logger interfaces.Logger,
) *AggregationService {
	return &AggregationService{
		productRepo:         productRepo,
		variantRepo:         variantRepo,
		inventoryRepo:       inventoryRepo,
		warehouseRepo:       warehouseRepo,
		collaboratorRepo:    collaboratorRepo,
		discountRepo:        discountRepo,
		taxRepo:             taxRepo,
		refundPoliciesRepo:  refundPoliciesRepo,
		logger:              logger,
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

	response := &models.ProductDetailResponse{
		Product: models.ProductInfo{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			IsActive:    true, // Products don't have IsActive field in this model
			CreatedAt:   product.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
		},
		Variants: []models.VariantDetail{},
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
						response.Collaborator = &models.CollaboratorInfo{
							ID:            collaborator.ID,
							CompanyName:   collaborator.CompanyName,
							ContactPerson: &collaborator.ContactPerson,
							Phone:         &collaborator.ContactNumber,
							Email:         collaborator.Email,
							IsActive:      isActive,
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
func (s *AggregationService) buildVariantDetail(v *models.ProductVariant, req *models.ProductDetailRequest, includes models.IncludeOptions) models.VariantDetail {
	detail := models.VariantDetail{
		ID:          v.ID,
		ProductID:   v.ProductID,
		VariantName: v.VariantName,
		Description: v.Description,
		Quantity:    v.Quantity,
		PackSize:    &v.PackSize,
		BrandName:   v.BrandName,
		HSNCode:     v.HSNCode,
		IsActive:    v.IsActive,
		CreatedAt:   v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   v.UpdatedAt.Format(time.RFC3339),
	}

	if v.SKU != nil {
		detail.SKU = *v.SKU
	}
	detail.Barcode = v.Barcode

	if v.ExternalID != nil {
		detail.ExternalID = v.ExternalID
	}

	if v.GSTRate != nil {
		detail.GSTRate = *v.GSTRate
	}

	detail.DosageInstructions = v.DosageInstructions
	detail.UsageDetails = v.UsageDetails

	// Parse images from JSON
	if v.Images != nil {
		var images []string
		if err := json.Unmarshal([]byte(*v.Images), &images); err == nil {
			detail.Images = images
		}
	}

	// Get prices if requested
	if includes.Prices && len(v.Prices) > 0 {
		prices := &models.VariantPrices{
			Currency:       "INR",
			HasActivePrice: false,
		}

		for _, p := range v.Prices {
			priceInfo := &models.PriceInfo{
				Price:         p.Price,
				EffectiveFrom: time.Now().Format(time.RFC3339), // Using embedded prices, no date range
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

		detail.Prices = prices
	}

	// Get inventory if requested
	if includes.Inventory {
		batches, err := s.inventoryRepo.GetBatchesByVariant(v.ID)
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

	// Get tax configuration if requested
	if includes.Taxes {
		// Get tax config from batch if available
		batches, err := s.inventoryRepo.GetBatchesByVariant(v.ID)
		if err == nil && len(batches) > 0 {
			// Use first batch's tax config
			batch := batches[0]
			detail.TaxConfiguration = &models.TaxConfiguration{
				CGSTRate:     batch.CGSTRate,
				SGSTRate:     batch.SGSTRate,
				IsTaxExempt:  batch.IsTaxExempt,
				CustomTaxIDs: batch.CustomTaxIDs,
			}
		} else if v.GSTRate != nil {
			// Fallback to variant GST rate
			gstHalf := *v.GSTRate / 2
			detail.TaxConfiguration = &models.TaxConfiguration{
				CGSTRate:     gstHalf,
				SGSTRate:     gstHalf,
				IsTaxExempt:  false,
				CustomTaxIDs: []string{},
			}
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
			VariantDetail: detail,
		},
		Metadata: models.ResponseMetadata{
			ReadTimestamp: time.Now().Format(time.RFC3339),
		},
	}

	// Include product info if requested
	if includes.Variants {
		product, err := s.productRepo.GetByID(variant.ProductID)
		if err == nil {
			response.Variant.Product = &models.ProductBasicInfo{
				ID:   product.ID,
				Name: product.Name,
				// Category not available in Product model
			}
		}
	}

	// Include collaborator info if requested
	if includes.Collaborators && len(variant.CollaboratorIDs) > 0 {
		collaboratorID := variant.CollaboratorIDs[0]
		collaborator, err := s.collaboratorRepo.GetByID(collaboratorID)
		if err == nil {
			response.Variant.Collaborator = &models.CollaboratorBasicInfo{
				ID:          collaborator.ID,
				CompanyName: collaborator.CompanyName,
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

		response.Warehouse = models.WarehouseInfo{
			ID:       warehouse.ID,
			Name:     warehouse.Name,
			IsActive: true, // Warehouse model doesn't have IsActive field, defaulting to true
		}
	}

	// Get inventory batches
	var batches []models.InventoryBatch
	var err error
	if req.WarehouseID != "" {
		batches, err = s.inventoryRepo.GetBatchesByWarehouse(req.WarehouseID)
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

		inventoryItem := models.InventoryWithPricing{
			BatchID:           batch.ID,
			VariantID:         batch.VariantID,
			QuantityAvailable: batch.TotalQuantity,
			QuantityReserved:  batch.ReservedQuantity,
			QuantitySellable:  batch.AvailableQuantity(), // Real sellable = total - reserved
			CostPrice:         batch.CostPrice,
			ExpiryDate:        batch.ExpiryDate.Format("2006-01-02"),
			TaxConfig: models.BatchTaxConfig{
				CGSTRate:     batch.CGSTRate,
				SGSTRate:     batch.SGSTRate,
				TotalGSTRate: batch.CGSTRate + batch.SGSTRate,
				IsTaxExempt:  batch.IsTaxExempt,
				CustomTaxes:  batch.CustomTaxIDs,
			},
		}

		// Get variant info
		variant, err := s.variantRepo.GetByID(batch.VariantID)
		if err == nil {
			inventoryItem.Variant = models.VariantInfoForSales{
				ID:          variant.ID,
				VariantName: variant.VariantName,
				Quantity:    variant.Quantity,
				PackSize:    &variant.PackSize,
				BrandName:   variant.BrandName,
				HSNCode:     variant.HSNCode,
				IsActive:    variant.IsActive,
			}
			if variant.SKU != nil {
				inventoryItem.Variant.SKU = *variant.SKU
			}
			inventoryItem.Variant.Barcode = variant.Barcode

			// Parse images
			if variant.Images != nil {
				var images []string
				if jsonErr := json.Unmarshal([]byte(*variant.Images), &images); jsonErr == nil {
					inventoryItem.Variant.Images = images
				}
			}

			// Get selling price from variant
			if len(variant.Prices) > 0 {
				priceType := strings.ToLower(req.PriceType)
				if priceType == "" {
					priceType = "retail"
				}

				for _, p := range variant.Prices {
					pType := strings.ToLower(p.PriceType)
					if pType == priceType || (priceType == "retail" && pType == "mrp") ||
						(priceType == "wholesale" && pType == "msp") {
						inventoryItem.SellingPrice = &models.SellingPriceInfo{
							Price:         p.Price,
							PriceType:     p.PriceType,
							Currency:      p.Currency,
							EffectiveFrom: time.Now().Format(time.RFC3339),
							IsActive:      true,
						}
						break
					}
				}

				// Build alternate prices
				for _, p := range variant.Prices {
					pType := strings.ToLower(p.PriceType)
					if inventoryItem.SellingPrice == nil || pType != strings.ToLower(inventoryItem.SellingPrice.PriceType) {
						inventoryItem.AlternatePrices = append(inventoryItem.AlternatePrices, models.AlternatePriceInfo{
							Price:     p.Price,
							PriceType: p.PriceType,
						})
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
				inventoryItem.Product = models.ProductInfoForSales{
					ID:          product.ID,
					Name:        product.Name,
					Description: product.Description,
					// Category not available in Product model
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

	// Get active taxes for global config
	taxes, err := s.taxRepo.GetTaxesByStatus("active")
	if err == nil {
		response.GlobalTaxConfiguration.TaxCalculationMethod = "exclusive"
		for _, t := range taxes {
			response.GlobalTaxConfiguration.ActiveTaxes = append(response.GlobalTaxConfiguration.ActiveTaxes, models.ActiveTaxInfo{
				ID:       t.ID,
				Name:     t.Name,
				TaxType:  string(t.TaxType),
				CGSTRate: t.Rate / 2, // Assuming GST split
				SGSTRate: t.Rate / 2,
				IsActive: t.IsActive,
			})

			// Set default rates from first GST tax
			if t.TaxType == models.TaxTypeCGST || t.TaxType == models.TaxTypeSGST {
				response.GlobalTaxConfiguration.DefaultCGSTRate = t.Rate
			}
		}
	}

	// Get refund policies
	refundPolicies, err := s.refundPoliciesRepo.GetAllRefundPolicies(100, 0)
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
