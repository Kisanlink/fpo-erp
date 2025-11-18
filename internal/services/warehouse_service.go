package services

import (
	"context"
	"fmt"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
)

// WarehouseService handles warehouse business logic
type WarehouseService struct {
	warehouseRepo *repositories.WarehouseRepository
	addressClient *aaa.AddressGRPCClient
}

// NewWarehouseService creates a new warehouse service
func NewWarehouseService(warehouseRepo *repositories.WarehouseRepository, addressClient *aaa.AddressGRPCClient) *WarehouseService {
	return &WarehouseService{
		warehouseRepo: warehouseRepo,
		addressClient: addressClient,
	}
}

// CreateWarehouse creates a new warehouse
func (s *WarehouseService) CreateWarehouse(ctx context.Context, request *models.CreateWarehouseRequest, userID string, jwtToken string) (*models.WarehouseResponse, error) {
	var addressID *string

	// Handle inline address creation if provided
	if request.Address != nil {
		// Create address via AAA service with timeout
		ctxAddr, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		// Use the authenticated user ID passed from handler
		address, err := s.addressClient.CreateAddress(ctxAddr, &aaa.CreateAddressRequest{
			UserID:      userID,
			Type:        request.Address.Type,
			House:       request.Address.House,
			Street:      request.Address.Street,
			Landmark:    request.Address.Landmark,
			PostOffice:  request.Address.PostOffice,
			Subdistrict: request.Address.Subdistrict,
			District:    request.Address.District,
			VTC:         request.Address.VTC,
			State:       request.Address.State,
			Country:     request.Address.Country,
			Pincode:     request.Address.Pincode,
			IsPrimary:   request.Address.IsPrimary,
		}, jwtToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create address: %w", err)
		}
		addressID = &address.ID
	} else if request.AddressID != nil {
		addressID = request.AddressID
	}

	// Create warehouse model using the proper constructor
	warehouse := models.NewWarehouse(request.Name, addressID)

	// Save to database
	if err := s.warehouseRepo.Create(warehouse); err != nil {
		if addressID != nil {
			_ = s.addressClient.DeleteAddress(ctx, *addressID, true, jwtToken) // best-effort rollback
		}
		return nil, err
	}

	// Build response with address details
	return s.buildWarehouseResponse(ctx, warehouse, jwtToken)
}

// GetWarehouse retrieves a warehouse by ID
func (s *WarehouseService) GetWarehouse(ctx context.Context, id string, jwtToken string) (*models.WarehouseResponse, error) {
	warehouse, err := s.warehouseRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.buildWarehouseResponse(ctx, warehouse, jwtToken)
}

// GetAllWarehouses retrieves all warehouses
func (s *WarehouseService) GetAllWarehouses(ctx context.Context, jwtToken string) ([]models.WarehouseResponse, error) {
	warehouses, err := s.warehouseRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []models.WarehouseResponse
	for _, warehouse := range warehouses {
		response, err := s.buildWarehouseResponse(ctx, &warehouse, jwtToken)
		if err != nil {
			// Log error but continue with other warehouses
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// UpdateWarehouse updates a warehouse
func (s *WarehouseService) UpdateWarehouse(ctx context.Context, id string, request *models.UpdateWarehouseRequest, jwtToken string) (*models.WarehouseResponse, error) {
	// Get existing warehouse
	warehouse, err := s.warehouseRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Handle inline address updates if provided
	if request.Address != nil {
		// Validate ownership/association before update
		if warehouse.AddressID == nil || request.Address.ID == "" || *warehouse.AddressID != request.Address.ID {
			return nil, errors.NewBadRequestError("address mismatch: update not permitted")
		}
		// Update address via AAA service
		address, err := s.addressClient.UpdateAddress(ctx, &aaa.UpdateAddressRequest{
			ID:          request.Address.ID,
			Type:        request.Address.Type,
			House:       request.Address.House,
			Street:      request.Address.Street,
			Landmark:    request.Address.Landmark,
			PostOffice:  request.Address.PostOffice,
			Subdistrict: request.Address.Subdistrict,
			District:    request.Address.District,
			VTC:         request.Address.VTC,
			State:       request.Address.State,
			Country:     request.Address.Country,
			Pincode:     request.Address.Pincode,
			IsPrimary:   request.Address.IsPrimary != nil && *request.Address.IsPrimary,
			IsActive:    true,
		}, jwtToken)
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

	return s.buildWarehouseResponse(ctx, warehouse, jwtToken)
}

// DeleteWarehouse deletes a warehouse
func (s *WarehouseService) DeleteWarehouse(ctx context.Context, id string, jwtToken string) error {
	// Get warehouse to check if it has an address
	warehouse, err := s.warehouseRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete associated address if exists
	if warehouse.AddressID != nil {
		if err := s.addressClient.DeleteAddress(ctx, *warehouse.AddressID, true, jwtToken); err != nil {
			// Log error but don't fail the warehouse deletion
			// You might want to handle this differently based on requirements
		}
	}

	// Delete warehouse
	return s.warehouseRepo.Delete(id)
}

// SearchWarehouses searches warehouses by name
func (s *WarehouseService) SearchWarehouses(ctx context.Context, query string, jwtToken string) ([]models.WarehouseResponse, error) {
	warehouses, err := s.warehouseRepo.GetByName(query)
	if err != nil {
		return nil, err
	}

	var responses []models.WarehouseResponse
	for _, warehouse := range warehouses {
		response, err := s.buildWarehouseResponse(ctx, &warehouse, jwtToken)
		if err != nil {
			// Log error but continue with other warehouses
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// buildWarehouseResponse builds a warehouse response with address details
func (s *WarehouseService) buildWarehouseResponse(ctx context.Context, warehouse *models.Warehouse, jwtToken string) (*models.WarehouseResponse, error) {
	response := &models.WarehouseResponse{
		ID:        warehouse.ID,
		Name:      warehouse.Name,
		CreatedAt: warehouse.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: warehouse.UpdatedAt.UTC().Format(time.RFC3339),
	}

	// Fetch address details if address ID exists
	if warehouse.AddressID != nil {
		address, err := s.addressClient.GetAddress(ctx, *warehouse.AddressID, jwtToken)
		if err != nil {
			// Log error but don't fail the request
			// You might want to handle this differently based on requirements
			return response, nil
		}

		response.Address = &models.AddressInfo{
			ID:          address.ID,
			Type:        address.Type,
			House:       address.House,
			Street:      address.Street,
			Landmark:    address.Landmark,
			PostOffice:  address.PostOffice,
			Subdistrict: address.Subdistrict,
			District:    address.District,
			VTC:         address.VTC,
			State:       address.State,
			Country:     address.Country,
			Pincode:     address.Pincode,
			FullAddress: address.BuildFullAddress(),
		}
	}

	return response, nil
}
