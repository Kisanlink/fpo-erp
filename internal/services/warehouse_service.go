package services

import (
	"context"
	"fmt"
	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
)

// WarehouseService handles warehouse business logic
type WarehouseService struct {
	warehouseRepo *repositories.WarehouseRepository
	addressClient *aaa.AddressClient
}

// NewWarehouseService creates a new warehouse service
func NewWarehouseService(warehouseRepo *repositories.WarehouseRepository, addressClient *aaa.AddressClient) *WarehouseService {
	return &WarehouseService{
		warehouseRepo: warehouseRepo,
		addressClient: addressClient,
	}
}

// CreateWarehouse creates a new warehouse
func (s *WarehouseService) CreateWarehouse(ctx context.Context, request *models.CreateWarehouseRequest) (*models.WarehouseResponse, error) {
	var addressID *string

	// Handle inline address creation if provided
	if request.Address != nil {
		// TODO: Implement address creation via AAA service when AddressService is available
		// For now, we'll skip address creation and store address data directly in warehouse
		// address, err := s.addressClient.CreateAddress(ctx, &aaa.CreateAddressRequest{
		// 	UserID:       "system", // or get from context
		// 	Type:         request.Address.Type,
		// 	AddressLine1: request.Address.AddressLine1,
		// 	AddressLine2: request.Address.AddressLine2,
		// 	City:         request.Address.City,
		// 	State:        request.Address.State,
		// 	PostalCode:   request.Address.PostalCode,
		// 	Country:      request.Address.Country,
		// 	IsPrimary:    request.Address.IsPrimary,
		// })
		// if err != nil {
		// 	return nil, fmt.Errorf("failed to create address: %w", err)
		// }
		// addressID = &address.ID

		// For now, set addressID to nil - address data will be stored in warehouse
		addressID = nil
	} else if request.AddressID != nil {
		addressID = request.AddressID
	}

	// Create warehouse model using the proper constructor
	var warehouse *models.Warehouse
	if request.Address != nil {
		// Use constructor with direct address fields
		warehouse = models.NewWarehouseWithAddress(request.Name, request.Address)
	} else {
		// Use constructor with address ID
		warehouse = models.NewWarehouse(request.Name, addressID)
	}

	// Save to database
	if err := s.warehouseRepo.Create(warehouse); err != nil {
		return nil, err
	}

	// Build response with address details
	return s.buildWarehouseResponse(ctx, warehouse)
}

// GetWarehouse retrieves a warehouse by ID
func (s *WarehouseService) GetWarehouse(ctx context.Context, id string) (*models.WarehouseResponse, error) {
	warehouse, err := s.warehouseRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.buildWarehouseResponse(ctx, warehouse)
}

// GetAllWarehouses retrieves all warehouses
func (s *WarehouseService) GetAllWarehouses(ctx context.Context) ([]models.WarehouseResponse, error) {
	warehouses, err := s.warehouseRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []models.WarehouseResponse
	for _, warehouse := range warehouses {
		response, err := s.buildWarehouseResponse(ctx, &warehouse)
		if err != nil {
			// Log error but continue with other warehouses
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// UpdateWarehouse updates a warehouse
func (s *WarehouseService) UpdateWarehouse(ctx context.Context, id string, request *models.UpdateWarehouseRequest) (*models.WarehouseResponse, error) {
	// Get existing warehouse
	warehouse, err := s.warehouseRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Handle inline address updates if provided
	if request.Address != nil {
		address, err := s.addressClient.UpdateAddress(ctx, &aaa.UpdateAddressRequest{
			ID:           request.Address.ID,
			Type:         request.Address.Type,
			AddressLine1: request.Address.AddressLine1,
			AddressLine2: request.Address.AddressLine2,
			City:         request.Address.City,
			State:        request.Address.State,
			PostalCode:   request.Address.PostalCode,
			Country:      request.Address.Country,
			IsPrimary:    request.Address.IsPrimary != nil && *request.Address.IsPrimary,
			IsActive:     true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update address: %w", err)
		}
		warehouse.AddressID = &address.ID
	}

	// Update fields if provided
	if request.Name != nil {
		warehouse.Name = *request.Name
	}
	if request.AddressID != nil {
		warehouse.AddressID = request.AddressID
	}

	// Save to database
	if err := s.warehouseRepo.Update(warehouse); err != nil {
		return nil, err
	}

	return s.buildWarehouseResponse(ctx, warehouse)
}

// DeleteWarehouse deletes a warehouse
func (s *WarehouseService) DeleteWarehouse(ctx context.Context, id string) error {
	// Get warehouse to check if it has an address
	warehouse, err := s.warehouseRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete associated address if exists
	if warehouse.AddressID != nil {
		if err := s.addressClient.DeleteAddress(ctx, *warehouse.AddressID, true); err != nil {
			// Log error but don't fail the warehouse deletion
			// You might want to handle this differently based on requirements
		}
	}

	// Delete warehouse
	return s.warehouseRepo.Delete(id)
}

// SearchWarehouses searches warehouses by name
func (s *WarehouseService) SearchWarehouses(ctx context.Context, query string) ([]models.WarehouseResponse, error) {
	warehouses, err := s.warehouseRepo.GetByName(query)
	if err != nil {
		return nil, err
	}

	var responses []models.WarehouseResponse
	for _, warehouse := range warehouses {
		response, err := s.buildWarehouseResponse(ctx, &warehouse)
		if err != nil {
			// Log error but continue with other warehouses
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// buildWarehouseResponse builds a warehouse response with address details
func (s *WarehouseService) buildWarehouseResponse(ctx context.Context, warehouse *models.Warehouse) (*models.WarehouseResponse, error) {
	response := &models.WarehouseResponse{
		ID:        warehouse.ID,
		Name:      warehouse.Name,
		CreatedAt: warehouse.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: warehouse.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Fetch address details if address ID exists
	if warehouse.AddressID != nil {
		address, err := s.addressClient.GetAddress(ctx, *warehouse.AddressID)
		if err != nil {
			// Log error but don't fail the request
			// You might want to handle this differently based on requirements
			return response, nil
		}

		response.Address = &models.AddressInfo{
			ID:           address.ID,
			Type:         address.Type,
			AddressLine1: address.AddressLine1,
			AddressLine2: address.AddressLine2,
			City:         address.City,
			State:        address.State,
			PostalCode:   address.PostalCode,
			Country:      address.Country,
			FullAddress:  s.buildFullAddress(address),
		}
	}

	return response, nil
}

// buildFullAddress builds a full address string from address components
func (s *WarehouseService) buildFullAddress(address *aaa.Address) string {
	parts := []string{}
	if address.AddressLine1 != "" {
		parts = append(parts, address.AddressLine1)
	}
	if address.AddressLine2 != "" {
		parts = append(parts, address.AddressLine2)
	}
	if address.City != "" {
		parts = append(parts, address.City)
	}
	if address.State != "" {
		parts = append(parts, address.State)
	}
	if address.PostalCode != "" {
		parts = append(parts, address.PostalCode)
	}
	if address.Country != "" {
		parts = append(parts, address.Country)
	}

	fullAddress := ""
	for i, part := range parts {
		if i > 0 {
			fullAddress += ", "
		}
		fullAddress += part
	}
	return fullAddress
}
