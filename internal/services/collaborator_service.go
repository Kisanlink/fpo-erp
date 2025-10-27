package services

import (
	"context"
	"fmt"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
)

// CollaboratorService handles collaborator business logic
type CollaboratorService struct {
	collaboratorRepo *repositories.CollaboratorRepository
	addressClient    *aaa.AddressClient
	s3Service        *S3Service
}

// NewCollaboratorService creates a new collaborator service
func NewCollaboratorService(collaboratorRepo *repositories.CollaboratorRepository, addressClient *aaa.AddressClient, s3Service *S3Service) *CollaboratorService {
	return &CollaboratorService{
		collaboratorRepo: collaboratorRepo,
		addressClient:    addressClient,
		s3Service:        s3Service,
	}
}

// CreateCollaborator creates a new collaborator with address
func (s *CollaboratorService) CreateCollaborator(ctx context.Context, request *models.CreateCollaboratorRequest) (*models.CollaboratorResponse, error) {
	var addressID *string

	// Handle inline address creation if provided
	if request.Address != nil {
		// Create address via AAA service with timeout
		ctxAddr, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		userID := "system" // TODO: Extract from auth context when available
		address, err := s.addressClient.CreateAddress(ctxAddr, &aaa.CreateAddressRequest{
			UserID:       userID,
			Type:         request.Address.Type,
			AddressLine1: request.Address.AddressLine1,
			AddressLine2: request.Address.AddressLine2,
			City:         request.Address.City,
			State:        request.Address.State,
			PostalCode:   request.Address.PostalCode,
			Country:      request.Address.Country,
			IsPrimary:    request.Address.IsPrimary,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create address: %w", err)
		}
		addressID = &address.ID
	}

	// Validate GST number doesn't already exist
	if request.GSTNumber != "" {
		exists, err := s.collaboratorRepo.GSTNumberExists(request.GSTNumber)
		if err != nil {
			return nil, err
		}
		if exists {
			// Rollback address creation if GST exists
			if addressID != nil {
				_ = s.addressClient.DeleteAddress(ctx, *addressID, true)
			}
			return nil, fmt.Errorf("collaborator with GST number %s already exists", request.GSTNumber)
		}
	}

	// Create collaborator model using the proper constructor
	collaborator := models.NewCollaborator(
		request.CompanyName,
		request.ContactPerson,
		request.ContactNumber,
		request.BankAccountNo,
		request.BankIFSC,
		addressID,
	)

	// Set optional fields
	collaborator.Logo = request.Logo
	collaborator.Email = request.Email
	collaborator.GSTNumber = request.GSTNumber
	collaborator.PANNumber = request.PANNumber
	collaborator.BankName = request.BankName
	collaborator.Experience = request.Experience

	// Save to database
	if err := s.collaboratorRepo.Create(collaborator); err != nil {
		if addressID != nil {
			_ = s.addressClient.DeleteAddress(ctx, *addressID, true) // best-effort rollback
		}
		return nil, err
	}

	// Build response with address details
	return s.buildCollaboratorResponse(ctx, collaborator)
}

// GetCollaborator retrieves a collaborator by ID
func (s *CollaboratorService) GetCollaborator(ctx context.Context, id string) (*models.CollaboratorResponse, error) {
	collaborator, err := s.collaboratorRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.buildCollaboratorResponse(ctx, collaborator)
}

// GetAllCollaborators retrieves all collaborators
func (s *CollaboratorService) GetAllCollaborators(ctx context.Context) ([]models.CollaboratorResponse, error) {
	collaborators, err := s.collaboratorRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []models.CollaboratorResponse
	for _, collaborator := range collaborators {
		response, err := s.buildCollaboratorResponse(ctx, &collaborator)
		if err != nil {
			// Log error but continue with other collaborators
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// GetActiveCollaborators retrieves all active collaborators
func (s *CollaboratorService) GetActiveCollaborators(ctx context.Context) ([]models.CollaboratorResponse, error) {
	collaborators, err := s.collaboratorRepo.GetActiveCollaborators()
	if err != nil {
		return nil, err
	}

	var responses []models.CollaboratorResponse
	for _, collaborator := range collaborators {
		response, err := s.buildCollaboratorResponse(ctx, &collaborator)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// UpdateCollaborator updates a collaborator
func (s *CollaboratorService) UpdateCollaborator(ctx context.Context, id string, request *models.UpdateCollaboratorRequest) (*models.CollaboratorResponse, error) {
	// Get existing collaborator
	collaborator, err := s.collaboratorRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Handle inline address updates if provided
	if request.Address != nil && collaborator.AddressID != nil {
		// Validate ownership/association before update
		if request.Address.ID == "" || *collaborator.AddressID != request.Address.ID {
			return nil, fmt.Errorf("address mismatch: update not permitted")
		}
		// Update address via AAA service
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
		collaborator.AddressID = &address.ID
	}

	// Update fields if provided
	if request.CompanyName != nil {
		collaborator.CompanyName = *request.CompanyName
	}
	if request.Logo != nil {
		collaborator.Logo = request.Logo
	}
	if request.ContactPerson != nil {
		collaborator.ContactPerson = *request.ContactPerson
	}
	if request.ContactNumber != nil {
		collaborator.ContactNumber = *request.ContactNumber
	}
	if request.Email != nil {
		collaborator.Email = request.Email
	}
	if request.GSTNumber != nil {
		collaborator.GSTNumber = *request.GSTNumber
	}
	if request.PANNumber != nil {
		collaborator.PANNumber = request.PANNumber
	}
	if request.BankAccountNo != nil {
		collaborator.BankAccountNo = *request.BankAccountNo
	}
	if request.BankIFSC != nil {
		collaborator.BankIFSC = *request.BankIFSC
	}
	if request.BankName != nil {
		collaborator.BankName = request.BankName
	}
	if request.Experience != nil {
		collaborator.Experience = request.Experience
	}
	if request.IsActive != nil {
		collaborator.IsActive = *request.IsActive
	}

	// Save to database
	if err := s.collaboratorRepo.Update(collaborator); err != nil {
		return nil, err
	}

	return s.buildCollaboratorResponse(ctx, collaborator)
}

// DeleteCollaborator deletes a collaborator (soft delete)
func (s *CollaboratorService) DeleteCollaborator(ctx context.Context, id string) error {
	// Get collaborator to check if it has an address
	collaborator, err := s.collaboratorRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete associated address if exists (soft delete)
	if collaborator.AddressID != nil {
		if err := s.addressClient.DeleteAddress(ctx, *collaborator.AddressID, true); err != nil {
			// Log error but don't fail the collaborator deletion
		}
	}

	// Delete collaborator (soft delete)
	return s.collaboratorRepo.Delete(id)
}

// SearchCollaborators searches collaborators by name
func (s *CollaboratorService) SearchCollaborators(ctx context.Context, query string) ([]models.CollaboratorResponse, error) {
	collaborators, err := s.collaboratorRepo.SearchByName(query)
	if err != nil {
		return nil, err
	}

	var responses []models.CollaboratorResponse
	for _, collaborator := range collaborators {
		response, err := s.buildCollaboratorResponse(ctx, &collaborator)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// buildCollaboratorResponse builds a collaborator response with address details
func (s *CollaboratorService) buildCollaboratorResponse(ctx context.Context, collaborator *models.Collaborator) (*models.CollaboratorResponse, error) {
	response := &models.CollaboratorResponse{
		ID:            collaborator.ID,
		CompanyName:   collaborator.CompanyName,
		Logo:          collaborator.Logo,
		ContactPerson: collaborator.ContactPerson,
		ContactNumber: collaborator.ContactNumber,
		Email:         collaborator.Email,
		GSTNumber:     collaborator.GSTNumber,
		PANNumber:     collaborator.PANNumber,
		BankAccountNo: collaborator.BankAccountNo,
		BankIFSC:      collaborator.BankIFSC,
		BankName:      collaborator.BankName,
		Experience:    collaborator.Experience,
		IsActive:      collaborator.IsActive,
		CreatedAt:     collaborator.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     collaborator.UpdatedAt.UTC().Format(time.RFC3339),
	}

	// Fetch address details if address ID exists
	if collaborator.AddressID != nil {
		address, err := s.addressClient.GetAddress(ctx, *collaborator.AddressID)
		if err != nil {
			// Log error but don't fail the request
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
func (s *CollaboratorService) buildFullAddress(address *aaa.Address) string {
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
