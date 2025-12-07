package services

import (
	"context"
	"fmt"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
)

// WarehouseService handles warehouse business logic
type WarehouseService struct {
	warehouseRepo *repositories.WarehouseRepository
	addressClient *aaa.AddressGRPCClient
	logger        interfaces.Logger
}

// NewWarehouseService creates a new warehouse service
func NewWarehouseService(warehouseRepo *repositories.WarehouseRepository, addressClient *aaa.AddressGRPCClient, logger interfaces.Logger) *WarehouseService {
	return &WarehouseService{
		warehouseRepo: warehouseRepo,
		addressClient: addressClient,
		logger:        logger,
	}
}

// CreateWarehouse creates a new warehouse
func (s *WarehouseService) CreateWarehouse(ctx context.Context, request *models.CreateWarehouseRequest, userID string, jwtToken string) (*models.WarehouseResponse, error) {
	s.logger.Info("Creating warehouse",
		zap.String("name", request.Name),
		zap.String("user_id", userID),
		zap.Bool("has_inline_address", request.Address != nil))

	// Check if AAA service is required but not available
	if s.addressClient == nil {
		s.logger.Error("AAA address service is not available - cannot create warehouse")
		return nil, errors.NewServiceUnavailableError("AAA address service is not available")
	}

	var addressID *string

	// Handle inline address creation if provided
	if request.Address != nil {
		state := ""
		if request.Address.State != nil {
			state = *request.Address.State
		}
		district := ""
		if request.Address.District != nil {
			district = *request.Address.District
		}
		s.logger.Debug("Creating address via AAA service",
			zap.String("state", state),
			zap.String("district", district))

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
			s.logger.Error("Failed to create address via AAA",
				zap.Error(err),
				zap.String("user_id", userID))
			return nil, fmt.Errorf("failed to create address: %w", err)
		}
		addressID = &address.ID
		s.logger.Debug("Address created successfully",
			zap.String("address_id", address.ID))
	} else if request.AddressID != nil {
		addressID = request.AddressID
		s.logger.Debug("Using existing address",
			zap.String("address_id", *addressID))
	}

	// Create warehouse model using the proper constructor
	warehouse := models.NewWarehouse(request.Name, addressID)

	s.logger.Debug("Saving warehouse to database")

	// Save to database
	if err := s.warehouseRepo.Create(warehouse); err != nil {
		s.logger.Error("Failed to create warehouse",
			zap.Error(err),
			zap.String("name", request.Name))
		if addressID != nil {
			s.logger.Warn("Rolling back address creation",
				zap.String("address_id", *addressID))
			_ = s.addressClient.DeleteAddress(ctx, *addressID, true, jwtToken) // best-effort rollback
		}
		return nil, err
	}

	s.logger.Info("Warehouse created successfully",
		zap.String("warehouse_id", warehouse.ID),
		zap.String("name", warehouse.Name))

	// Build response with address details
	return s.buildWarehouseResponse(ctx, warehouse, jwtToken)
}

// GetWarehouse retrieves a warehouse by ID
func (s *WarehouseService) GetWarehouse(ctx context.Context, id string, jwtToken string) (*models.WarehouseResponse, error) {
	s.logger.Info("Retrieving warehouse",
		zap.String("warehouse_id", id))

	// Check if AAA service is available
	if s.addressClient == nil {
		s.logger.Error("AAA address service is not available - cannot retrieve warehouse")
		return nil, errors.NewServiceUnavailableError("AAA address service is not available")
	}

	warehouse, err := s.warehouseRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve warehouse",
			zap.Error(err),
			zap.String("warehouse_id", id))
		return nil, err
	}

	s.logger.Debug("Warehouse retrieved successfully",
		zap.String("warehouse_id", id),
		zap.String("name", warehouse.Name))

	return s.buildWarehouseResponse(ctx, warehouse, jwtToken)
}

// GetAllWarehouses retrieves all warehouses
func (s *WarehouseService) GetAllWarehouses(ctx context.Context, jwtToken string) ([]models.WarehouseResponse, error) {
	s.logger.Info("Retrieving all warehouses")

	// Check if AAA service is available
	if s.addressClient == nil {
		s.logger.Error("AAA address service is not available - cannot retrieve warehouses")
		return nil, errors.NewServiceUnavailableError("AAA address service is not available")
	}

	warehouses, err := s.warehouseRepo.GetAll()
	if err != nil {
		s.logger.Error("Failed to retrieve all warehouses",
			zap.Error(err))
		return nil, err
	}

	var responses []models.WarehouseResponse
	for _, warehouse := range warehouses {
		response, err := s.buildWarehouseResponse(ctx, &warehouse, jwtToken)
		if err != nil {
			// Log error but continue with other warehouses
			s.logger.Warn("Failed to build warehouse response",
				zap.Error(err),
				zap.String("warehouse_id", warehouse.ID))
			continue
		}
		responses = append(responses, *response)
	}

	s.logger.Info("Retrieved all warehouses successfully",
		zap.Int("count", len(responses)))

	return responses, nil
}

// UpdateWarehouse updates a warehouse
func (s *WarehouseService) UpdateWarehouse(ctx context.Context, id string, request *models.UpdateWarehouseRequest, jwtToken string) (*models.WarehouseResponse, error) {
	s.logger.Info("Updating warehouse",
		zap.String("warehouse_id", id))

	// Check if AAA service is available
	if s.addressClient == nil {
		s.logger.Error("AAA address service is not available - cannot update warehouse")
		return nil, errors.NewServiceUnavailableError("AAA address service is not available")
	}

	// Get existing warehouse
	warehouse, err := s.warehouseRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve warehouse for update",
			zap.Error(err),
			zap.String("warehouse_id", id))
		return nil, err
	}

	// Handle inline address updates if provided
	if request.Address != nil {
		s.logger.Debug("Updating warehouse address",
			zap.String("warehouse_id", id),
			zap.String("address_id", request.Address.ID))

		// Validate ownership/association before update
		if warehouse.AddressID == nil || request.Address.ID == "" || *warehouse.AddressID != request.Address.ID {
			s.logger.Warn("Address mismatch during warehouse update",
				zap.String("warehouse_id", id),
				zap.String("expected_address_id", func() string {
					if warehouse.AddressID != nil {
						return *warehouse.AddressID
					}
					return "nil"
				}()),
				zap.String("provided_address_id", request.Address.ID))
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
			s.logger.Error("Failed to update address via AAA",
				zap.Error(err),
				zap.String("address_id", request.Address.ID))
			return nil, fmt.Errorf("failed to update address: %w", err)
		}
		warehouse.AddressID = &address.ID
		s.logger.Debug("Address updated successfully",
			zap.String("address_id", address.ID))
	}

	s.logger.Debug("Applying warehouse updates",
		zap.String("warehouse_id", id))

	// Update fields if provided
	if request.Name != nil {
		warehouse.Name = *request.Name
	}
	if request.AddressID != nil {
		warehouse.AddressID = request.AddressID
	}

	// Save to database
	if err := s.warehouseRepo.Update(warehouse); err != nil {
		s.logger.Error("Failed to update warehouse",
			zap.Error(err),
			zap.String("warehouse_id", id))
		return nil, err
	}

	s.logger.Info("Warehouse updated successfully",
		zap.String("warehouse_id", id),
		zap.String("name", warehouse.Name))

	return s.buildWarehouseResponse(ctx, warehouse, jwtToken)
}

// DeleteWarehouse deletes a warehouse
func (s *WarehouseService) DeleteWarehouse(ctx context.Context, id string, jwtToken string) error {
	s.logger.Info("Deleting warehouse",
		zap.String("warehouse_id", id))

	// Check if AAA service is available
	if s.addressClient == nil {
		s.logger.Error("AAA address service is not available - cannot delete warehouse")
		return errors.NewServiceUnavailableError("AAA address service is not available")
	}

	// Get warehouse to check if it has an address
	warehouse, err := s.warehouseRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve warehouse for deletion",
			zap.Error(err),
			zap.String("warehouse_id", id))
		return err
	}

	// Delete associated address if exists
	if warehouse.AddressID != nil {
		s.logger.Debug("Deleting associated address",
			zap.String("warehouse_id", id),
			zap.String("address_id", *warehouse.AddressID))

		if err := s.addressClient.DeleteAddress(ctx, *warehouse.AddressID, true, jwtToken); err != nil {
			// Log error but don't fail the warehouse deletion
			s.logger.Warn("Failed to delete associated address",
				zap.Error(err),
				zap.String("address_id", *warehouse.AddressID))
		}
	}

	// Delete warehouse
	if err := s.warehouseRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete warehouse",
			zap.Error(err),
			zap.String("warehouse_id", id))
		return err
	}

	s.logger.Info("Warehouse deleted successfully",
		zap.String("warehouse_id", id))

	return nil
}

// SearchWarehouses searches warehouses by name
func (s *WarehouseService) SearchWarehouses(ctx context.Context, query string, jwtToken string) ([]models.WarehouseResponse, error) {
	s.logger.Info("Searching warehouses",
		zap.String("query", query))

	// Check if AAA service is available
	if s.addressClient == nil {
		s.logger.Error("AAA address service is not available - cannot search warehouses")
		return nil, errors.NewServiceUnavailableError("AAA address service is not available")
	}

	warehouses, err := s.warehouseRepo.GetByName(query)
	if err != nil {
		s.logger.Error("Failed to search warehouses",
			zap.Error(err),
			zap.String("query", query))
		return nil, err
	}

	var responses []models.WarehouseResponse
	for _, warehouse := range warehouses {
		response, err := s.buildWarehouseResponse(ctx, &warehouse, jwtToken)
		if err != nil {
			// Log error but continue with other warehouses
			s.logger.Warn("Failed to build warehouse response",
				zap.Error(err),
				zap.String("warehouse_id", warehouse.ID))
			continue
		}
		responses = append(responses, *response)
	}

	s.logger.Info("Warehouse search completed",
		zap.String("query", query),
		zap.Int("results", len(responses)))

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
