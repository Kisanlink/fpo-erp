package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kisanlink-erp/internal/aaa"
	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"

	pb "github.com/Kisanlink/kisanlink-ecom/proto/gen/go/collaborator/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type collaboratorSyncClient interface {
	CreateCollaborator(ctx context.Context, in *pb.CreateCollaboratorRequest, opts ...grpc.CallOption) (*pb.CollaboratorResponse, error)
	UpdateCollaborator(ctx context.Context, in *pb.UpdateCollaboratorRequest, opts ...grpc.CallOption) (*pb.CollaboratorResponse, error)
	DeactivateCollaborator(ctx context.Context, in *pb.DeactivateCollaboratorRequest, opts ...grpc.CallOption) (*pb.StatusResponse, error)
}

// CollaboratorService handles collaborator business logic
type CollaboratorService struct {
	collaboratorRepo *repositories.CollaboratorRepository
	addressClient    *aaa.AddressGRPCClient
	s3Service        *S3Service
	ecomClient       collaboratorSyncClient
	ecomTimeout      time.Duration
	ecomAuthToken    string
}

// NewCollaboratorService creates a new collaborator service
func NewCollaboratorService(
	collaboratorRepo *repositories.CollaboratorRepository,
	addressClient *aaa.AddressGRPCClient,
	s3Service *S3Service,
	ecomClient collaboratorSyncClient,
	ecomTimeout time.Duration,
	ecomAuthToken string,
) *CollaboratorService {
	if ecomTimeout <= 0 {
		ecomTimeout = 5 * time.Second
	}
	return &CollaboratorService{
		collaboratorRepo: collaboratorRepo,
		addressClient:    addressClient,
		s3Service:        s3Service,
		ecomClient:       ecomClient,
		ecomTimeout:      ecomTimeout,
		ecomAuthToken:    ecomAuthToken,
	}
}

// CreateCollaborator creates a new collaborator with address
func (s *CollaboratorService) CreateCollaborator(ctx context.Context, request *models.CreateCollaboratorRequest, organizationID string, userID string, jwtToken string) (*models.CollaboratorResponse, error) {
	if s.ecomClient == nil {
		return s.createCollaboratorLegacy(ctx, request, userID, jwtToken)
	}
	return s.createCollaboratorViaEcommerce(ctx, request, organizationID, userID, jwtToken)
}

// GetCollaborator retrieves a collaborator by ID
func (s *CollaboratorService) GetCollaborator(ctx context.Context, id string, jwtToken string) (*models.CollaboratorResponse, error) {
	collaborator, err := s.collaboratorRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.buildCollaboratorResponse(ctx, collaborator, jwtToken)
}

// GetAllCollaborators retrieves all collaborators
func (s *CollaboratorService) GetAllCollaborators(ctx context.Context, jwtToken string) ([]models.CollaboratorResponse, error) {
	collaborators, err := s.collaboratorRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []models.CollaboratorResponse
	for _, collaborator := range collaborators {
		response, err := s.buildCollaboratorResponse(ctx, &collaborator, jwtToken)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// GetActiveCollaborators retrieves all active collaborators
func (s *CollaboratorService) GetActiveCollaborators(ctx context.Context, jwtToken string) ([]models.CollaboratorResponse, error) {
	collaborators, err := s.collaboratorRepo.GetActiveCollaborators()
	if err != nil {
		return nil, err
	}

	var responses []models.CollaboratorResponse
	for _, collaborator := range collaborators {
		response, err := s.buildCollaboratorResponse(ctx, &collaborator, jwtToken)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// UpdateCollaborator updates a collaborator
func (s *CollaboratorService) UpdateCollaborator(ctx context.Context, id string, request *models.UpdateCollaboratorRequest, organizationID string, userID string, jwtToken string) (*models.CollaboratorResponse, error) {
	if s.ecomClient == nil {
		return s.updateCollaboratorLegacy(ctx, id, request, jwtToken)
	}
	return s.updateCollaboratorViaEcommerce(ctx, id, request, organizationID, userID, jwtToken)
}

// DeleteCollaborator deletes a collaborator (soft delete)
func (s *CollaboratorService) DeleteCollaborator(ctx context.Context, id string, organizationID string, jwtToken string) error {
	if s.ecomClient == nil {
		return s.deleteCollaboratorLegacy(ctx, id, jwtToken)
	}
	return s.deleteCollaboratorViaEcommerce(ctx, id, organizationID, jwtToken)
}

// SearchCollaborators searches collaborators by name
func (s *CollaboratorService) SearchCollaborators(ctx context.Context, query string, jwtToken string) ([]models.CollaboratorResponse, error) {
	collaborators, err := s.collaboratorRepo.SearchByName(query)
	if err != nil {
		return nil, err
	}

	var responses []models.CollaboratorResponse
	for _, collaborator := range collaborators {
		response, err := s.buildCollaboratorResponse(ctx, &collaborator, jwtToken)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// createCollaboratorViaEcommerce syncs collaborator creation with the e-commerce service
func (s *CollaboratorService) createCollaboratorViaEcommerce(ctx context.Context, request *models.CreateCollaboratorRequest, organizationID string, userID string, jwtToken string) (*models.CollaboratorResponse, error) {
	if request == nil {
		return nil, errors.NewValidationError("collaborator request cannot be nil")
	}

	email := stringValue(request.Email)
	if email == "" {
		return nil, errors.NewValidationError("email is required for collaborator creation")
	}

	pbReq, err := buildCreateCollaboratorRequest(request, organizationID, userID, email)
	if err != nil {
		return nil, err
	}

	ecomCtx, cancel := s.newEcommerceContext(ctx, organizationID, userID)
	defer cancel()

	resp, err := s.ecomClient.CreateCollaborator(ecomCtx, pbReq)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to sync collaborator to e-commerce: %s", grpcErrorMessage(err)))
	}

	remote := resp.GetCollaborator()
	if remote == nil {
		return nil, errors.NewInternalServerError("e-commerce collaborator response missing payload")
	}

	externalID := remote.GetId()
	if externalID == "" {
		return nil, errors.NewInternalServerError("e-commerce collaborator response missing id")
	}

	addressID := extractAddressID(remote)
	statusPtr := statusToBool(remote.GetStatus())

	existing, err := s.collaboratorRepo.FindByExternalID(externalID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		s.applyCreateRequestToModel(existing, request, addressID)
		existing.ExternalID = ptrString(externalID)
		if statusPtr != nil {
			existing.IsActive = statusPtr
		}
		if err := s.collaboratorRepo.Update(existing); err != nil {
			return nil, err
		}
		return s.buildCollaboratorResponse(ctx, existing, jwtToken)
	}

	collaborator := models.NewCollaborator(
		request.CompanyName,
		request.ContactPerson,
		request.ContactNumber,
		request.BankAccountNo,
		request.BankIFSC,
		addressID,
	)
	collaborator.ExternalID = ptrString(externalID)
	collaborator.Logo = request.Logo
	collaborator.Email = request.Email
	collaborator.GSTNumber = request.GSTNumber
	collaborator.PANNumber = request.PANNumber
	collaborator.BankName = request.BankName
	collaborator.Experience = request.Experience
	if statusPtr != nil {
		collaborator.IsActive = statusPtr
	}

	if err := s.collaboratorRepo.Create(collaborator); err != nil {
		return nil, err
	}

	return s.buildCollaboratorResponse(ctx, collaborator, jwtToken)
}

func (s *CollaboratorService) createCollaboratorLegacy(ctx context.Context, request *models.CreateCollaboratorRequest, userID string, jwtToken string) (*models.CollaboratorResponse, error) {
	var addressID *string

	if request.Address != nil {
		ctxAddr, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

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
			return nil, errors.NewInternalServerError("failed to create address")
		}
		addressID = &address.ID
	}

	if request.GSTNumber != "" {
		exists, err := s.collaboratorRepo.GSTNumberExists(request.GSTNumber)
		if err != nil {
			return nil, err
		}
		if exists {
			if addressID != nil {
				_ = s.addressClient.DeleteAddress(ctx, *addressID, true, jwtToken)
			}
			return nil, errors.NewConflictError(fmt.Sprintf("collaborator with GST number %s already exists", request.GSTNumber))
		}
	}

	collaborator := models.NewCollaborator(
		request.CompanyName,
		request.ContactPerson,
		request.ContactNumber,
		request.BankAccountNo,
		request.BankIFSC,
		addressID,
	)

	collaborator.Logo = request.Logo
	collaborator.Email = request.Email
	collaborator.GSTNumber = request.GSTNumber
	collaborator.PANNumber = request.PANNumber
	collaborator.BankName = request.BankName
	collaborator.Experience = request.Experience

	if err := s.collaboratorRepo.Create(collaborator); err != nil {
		if addressID != nil {
			_ = s.addressClient.DeleteAddress(ctx, *addressID, true, jwtToken)
		}
		return nil, err
	}

	return s.buildCollaboratorResponse(ctx, collaborator, jwtToken)
}

func (s *CollaboratorService) updateCollaboratorViaEcommerce(ctx context.Context, id string, request *models.UpdateCollaboratorRequest, organizationID string, userID string, jwtToken string) (*models.CollaboratorResponse, error) {
	if request == nil {
		return nil, errors.NewValidationError("update request cannot be nil")
	}

	collaborator, err := s.collaboratorRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if collaborator.ExternalID == nil || *collaborator.ExternalID == "" {
		return s.updateCollaboratorLegacy(ctx, id, request, jwtToken)
	}

	pbReq, err := buildUpdateCollaboratorRequest(collaborator, request, organizationID, userID)
	if err != nil {
		return nil, err
	}

	if pbReq != nil {
		ecomCtx, cancel := s.newEcommerceContext(ctx, organizationID, userID)
		defer cancel()

		resp, err := s.ecomClient.UpdateCollaborator(ecomCtx, pbReq)
		if err != nil {
			return nil, errors.NewInternalServerError(fmt.Sprintf("failed to update collaborator in e-commerce: %s", grpcErrorMessage(err)))
		}
		if resp != nil && resp.Collaborator != nil {
			if addrID := extractAddressID(resp.Collaborator); addrID != nil {
				collaborator.AddressID = addrID
			}
			if statusPtr := statusToBool(resp.Collaborator.GetStatus()); statusPtr != nil {
				collaborator.IsActive = statusPtr
			}
			if email := resp.Collaborator.GetEmail(); email != "" {
				collaborator.Email = ptrString(email)
			}
		}
	}

	if request.IsActive != nil && collaborator.ExternalID != nil && *collaborator.ExternalID != "" {
		statusValue := pb.CollaboratorStatus_COLLABORATOR_STATUS_ACTIVE
		if !*request.IsActive {
			statusValue = pb.CollaboratorStatus_COLLABORATOR_STATUS_INACTIVE
		}
		if err := s.syncCollaboratorStatus(ctx, *collaborator.ExternalID, statusValue, organizationID, userID); err != nil {
			return nil, err
		}
	}

	if request.Address != nil {
		if err := s.updateCollaboratorAddress(ctx, collaborator, request.Address, jwtToken); err != nil {
			return nil, err
		}
	}

	s.applyUpdateRequestToModel(collaborator, request)

	if err := s.collaboratorRepo.Update(collaborator); err != nil {
		return nil, err
	}

	return s.buildCollaboratorResponse(ctx, collaborator, jwtToken)
}

func (s *CollaboratorService) updateCollaboratorLegacy(ctx context.Context, id string, request *models.UpdateCollaboratorRequest, jwtToken string) (*models.CollaboratorResponse, error) {
	collaborator, err := s.collaboratorRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if request.Address != nil && collaborator.AddressID != nil {
		if request.Address.ID == "" || *collaborator.AddressID != request.Address.ID {
			return nil, errors.NewBadRequestError("address mismatch: update not permitted")
		}
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
			return nil, errors.NewInternalServerError("failed to update address")
		}
		collaborator.AddressID = &address.ID
	}

	s.applyUpdateRequestToModel(collaborator, request)

	if err := s.collaboratorRepo.Update(collaborator); err != nil {
		return nil, err
	}

	return s.buildCollaboratorResponse(ctx, collaborator, jwtToken)
}

func (s *CollaboratorService) deleteCollaboratorViaEcommerce(ctx context.Context, id string, organizationID string, jwtToken string) error {
	collaborator, err := s.collaboratorRepo.GetByID(id)
	if err != nil {
		return err
	}

	if collaborator.ExternalID != nil && *collaborator.ExternalID != "" {
		if err := s.syncCollaboratorStatus(ctx, *collaborator.ExternalID, pb.CollaboratorStatus_COLLABORATOR_STATUS_INACTIVE, organizationID, ""); err != nil {
			return err
		}
	}

	return s.collaboratorRepo.Delete(id)
}

func (s *CollaboratorService) deleteCollaboratorLegacy(ctx context.Context, id string, jwtToken string) error {
	collaborator, err := s.collaboratorRepo.GetByID(id)
	if err != nil {
		return err
	}

	if collaborator.AddressID != nil {
		if err := s.addressClient.DeleteAddress(ctx, *collaborator.AddressID, true, jwtToken); err != nil {
			// ignore best-effort delete
		}
	}

	return s.collaboratorRepo.Delete(id)
}

// buildCollaboratorResponse builds a collaborator response with address details
func (s *CollaboratorService) buildCollaboratorResponse(ctx context.Context, collaborator *models.Collaborator, jwtToken string) (*models.CollaboratorResponse, error) {
	isActiveValue := false
	if collaborator.IsActive != nil {
		isActiveValue = *collaborator.IsActive
	}
	response := &models.CollaboratorResponse{
		ID:            collaborator.ID,
		ExternalID:    collaborator.ExternalID,
		AddressID:     collaborator.AddressID,
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
		IsActive:      isActiveValue,
		CreatedAt:     collaborator.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     collaborator.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if collaborator.AddressID != nil {
		address, err := s.addressClient.GetAddress(ctx, *collaborator.AddressID, jwtToken)
		if err == nil {
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
	}

	return response, nil
}

func (s *CollaboratorService) applyCreateRequestToModel(collaborator *models.Collaborator, request *models.CreateCollaboratorRequest, addressID *string) {
	collaborator.CompanyName = request.CompanyName
	collaborator.ContactPerson = request.ContactPerson
	collaborator.ContactNumber = request.ContactNumber
	collaborator.BankAccountNo = request.BankAccountNo
	collaborator.BankIFSC = request.BankIFSC
	collaborator.AddressID = addressID
	collaborator.Logo = request.Logo
	collaborator.Email = request.Email
	collaborator.GSTNumber = request.GSTNumber
	collaborator.PANNumber = request.PANNumber
	collaborator.BankName = request.BankName
	collaborator.Experience = request.Experience
}

func (s *CollaboratorService) applyUpdateRequestToModel(collaborator *models.Collaborator, request *models.UpdateCollaboratorRequest) {
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
		collaborator.IsActive = request.IsActive
	}
}

func (s *CollaboratorService) updateCollaboratorAddress(ctx context.Context, collaborator *models.Collaborator, req *models.UpdateAddressRequest, jwtToken string) error {
	if collaborator.AddressID == nil {
		return errors.NewBadRequestError("address update requested but collaborator has no address")
	}
	if req.ID == "" || *collaborator.AddressID != req.ID {
		return errors.NewBadRequestError("address mismatch: update not permitted")
	}
	_, err := s.addressClient.UpdateAddress(ctx, &aaa.UpdateAddressRequest{
		ID:          req.ID,
		Type:        req.Type,
		House:       req.House,
		Street:      req.Street,
		Landmark:    req.Landmark,
		PostOffice:  req.PostOffice,
		Subdistrict: req.Subdistrict,
		District:    req.District,
		VTC:         req.VTC,
		State:       req.State,
		Country:     req.Country,
		Pincode:     req.Pincode,
		IsPrimary:   req.IsPrimary != nil && *req.IsPrimary,
		IsActive:    true,
	}, jwtToken)
	return err
}

func (s *CollaboratorService) syncCollaboratorStatus(ctx context.Context, externalID string, newStatus pb.CollaboratorStatus, organizationID string, userID string) error {
	ecomCtx, cancel := s.newEcommerceContext(ctx, organizationID, userID)
	defer cancel()

	req := &pb.DeactivateCollaboratorRequest{
		Id:        externalID,
		NewStatus: &newStatus,
	}

	_, err := s.ecomClient.DeactivateCollaborator(ecomCtx, req)
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to update collaborator status in e-commerce: %s", grpcErrorMessage(err)))
	}
	return nil
}

func (s *CollaboratorService) newEcommerceContext(ctx context.Context, organizationID string, userID string) (context.Context, context.CancelFunc) {
	timeoutCtx, cancel := context.WithTimeout(ctx, s.ecomTimeout)

	var pairs []string
	if s.ecomAuthToken != "" {
		pairs = append(pairs, "authorization", "Bearer "+s.ecomAuthToken)
	}
	if organizationID != "" {
		pairs = append(pairs, "x-fpo-id", organizationID)
	}
	if userID != "" {
		pairs = append(pairs, "x-user-id", userID)
	}

	if len(pairs) == 0 {
		return timeoutCtx, cancel
	}

	md := metadata.Pairs(pairs...)
	return metadata.NewOutgoingContext(timeoutCtx, md), cancel
}

func buildCreateCollaboratorRequest(req *models.CreateCollaboratorRequest, organizationID string, userID string, email string) (*pb.CreateCollaboratorRequest, error) {
	firstName, lastName := splitContactName(req.ContactPerson)

	userIdentifier := userID
	if userIdentifier == "" {
		userIdentifier = req.ContactNumber
	}
	if userIdentifier == "" {
		userIdentifier = req.CompanyName
	}

	pbReq := &pb.CreateCollaboratorRequest{
		UserId:    userIdentifier,
		Username:  sanitizeUsername(req.ContactPerson),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Type:      pb.CollaboratorType_COLLABORATOR_TYPE_VENDOR,
	}
	if organizationID != "" {
		pbReq.OrganizationId = ptrString(organizationID)
	}

	businessInfo := &pb.CreateBusinessInfoRequest{
		BusinessName: req.CompanyName,
		BusinessType: pb.BusinessType_BUSINESS_TYPE_UNSPECIFIED,
	}
	if req.GSTNumber != "" {
		businessInfo.GstNumber = ptrString(req.GSTNumber)
	}
	if req.PANNumber != nil && *req.PANNumber != "" {
		businessInfo.PanNumber = req.PANNumber
	}
	if req.BankAccountNo != "" {
		businessInfo.BankAccountNumber = ptrString(req.BankAccountNo)
	}
	if req.BankIFSC != "" {
		businessInfo.BankIfscCode = ptrString(req.BankIFSC)
	}
	if req.BankName != nil && *req.BankName != "" {
		businessInfo.BankName = req.BankName
	}
	if req.ContactNumber != "" {
		businessInfo.BusinessPhone = ptrString(req.ContactNumber)
	}
	if req.Email != nil && *req.Email != "" {
		businessInfo.BusinessEmail = req.Email
	}
	pbReq.BusinessInfo = businessInfo

	if req.Address != nil {
		addressReq, err := buildCreateAddressRequest(req.Address)
		if err != nil {
			return nil, err
		}
		pbReq.Address = addressReq
	}

	return pbReq, nil
}

func buildUpdateCollaboratorRequest(existing *models.Collaborator, req *models.UpdateCollaboratorRequest, organizationID string, userID string) (*pb.UpdateCollaboratorRequest, error) {
	if existing.ExternalID == nil || *existing.ExternalID == "" {
		return nil, errors.NewValidationError("collaborator missing external id")
	}

	updateReq := &pb.UpdateCollaboratorRequest{
		Id: *existing.ExternalID,
	}

	var hasUpdates bool

	if req.Email != nil && *req.Email != "" {
		updateReq.Email = req.Email
		hasUpdates = true
	}
	if req.ContactNumber != nil && *req.ContactNumber != "" {
		updateReq.Phone = req.ContactNumber
		hasUpdates = true
	}
	if req.ContactPerson != nil && *req.ContactPerson != "" {
		firstName, lastName := splitContactName(*req.ContactPerson)
		updateReq.FirstName = ptrString(firstName)
		updateReq.LastName = ptrString(lastName)
		hasUpdates = true
	}

	var businessUpdate *pb.UpdateBusinessInfoRequest
	if req.CompanyName != nil && *req.CompanyName != "" {
		businessUpdate = ensureBusinessUpdate(businessUpdate)
		businessUpdate.BusinessName = req.CompanyName
		hasUpdates = true
	}
	if req.BankAccountNo != nil && *req.BankAccountNo != "" {
		businessUpdate = ensureBusinessUpdate(businessUpdate)
		businessUpdate.BankAccountNumber = req.BankAccountNo
		hasUpdates = true
	}
	if req.BankIFSC != nil && *req.BankIFSC != "" {
		businessUpdate = ensureBusinessUpdate(businessUpdate)
		businessUpdate.BankIfscCode = req.BankIFSC
		hasUpdates = true
	}
	if req.BankName != nil && *req.BankName != "" {
		businessUpdate = ensureBusinessUpdate(businessUpdate)
		businessUpdate.BankName = req.BankName
		hasUpdates = true
	}
	if req.ContactNumber != nil && *req.ContactNumber != "" {
		businessUpdate = ensureBusinessUpdate(businessUpdate)
		businessUpdate.BusinessPhone = req.ContactNumber
		hasUpdates = true
	}
	if req.Email != nil && *req.Email != "" {
		businessUpdate = ensureBusinessUpdate(businessUpdate)
		businessUpdate.BusinessEmail = req.Email
		hasUpdates = true
	}
	if businessUpdate != nil {
		updateReq.BusinessInfo = businessUpdate
	}

	if !hasUpdates {
		return nil, nil
	}

	return updateReq, nil
}

func buildCreateAddressRequest(req *models.CreateAddressRequest) (*pb.CreateAddressRequest, error) {
	line1Parts := []string{}
	if req.House != nil && *req.House != "" {
		line1Parts = append(line1Parts, *req.House)
	}
	if req.Street != nil && *req.Street != "" {
		line1Parts = append(line1Parts, *req.Street)
	}
	if len(line1Parts) == 0 && req.Landmark != nil && *req.Landmark != "" {
		line1Parts = append(line1Parts, *req.Landmark)
	}
	line1 := strings.TrimSpace(strings.Join(line1Parts, " "))
	if line1 == "" {
		line1 = "Address Line 1"
	}

	city := firstNonEmpty(req.VTC, req.District, req.Subdistrict)
	if city == "" {
		return nil, errors.NewValidationError("city (vtc/district/subdistrict) is required for address")
	}
	state := firstNonEmpty(req.State)
	if state == "" {
		return nil, errors.NewValidationError("state is required for address")
	}
	country := firstNonEmpty(req.Country)
	if country == "" {
		country = "India"
	}
	pincode := firstNonEmpty(req.Pincode)
	if pincode == "" {
		return nil, errors.NewValidationError("pincode is required for address")
	}

	address := &pb.CreateAddressRequest{
		Line1:      line1,
		City:       city,
		State:      state,
		Country:    country,
		PostalCode: pincode,
		Type:       mapAddressType(req.Type),
		IsPrimary:  req.IsPrimary,
	}
	if req.Landmark != nil && *req.Landmark != "" {
		address.Landmark = req.Landmark
	}
	return address, nil
}

func ensureBusinessUpdate(update *pb.UpdateBusinessInfoRequest) *pb.UpdateBusinessInfoRequest {
	if update == nil {
		update = &pb.UpdateBusinessInfoRequest{}
	}
	return update
}

func extractAddressID(collab *pb.Collaborator) *string {
	if collab == nil {
		return nil
	}
	if collab.PrimaryAddress != nil && collab.PrimaryAddress.Id != "" {
		return ptrString(collab.PrimaryAddress.Id)
	}
	if len(collab.AddressIds) > 0 && collab.AddressIds[0] != "" {
		return ptrString(collab.AddressIds[0])
	}
	return nil
}

func statusToBool(status pb.CollaboratorStatus) *bool {
	switch status {
	case pb.CollaboratorStatus_COLLABORATOR_STATUS_ACTIVE,
		pb.CollaboratorStatus_COLLABORATOR_STATUS_VERIFIED,
		pb.CollaboratorStatus_COLLABORATOR_STATUS_PENDING_VERIFICATION:
		return ptrBool(true)
	case pb.CollaboratorStatus_COLLABORATOR_STATUS_INACTIVE,
		pb.CollaboratorStatus_COLLABORATOR_STATUS_SUSPENDED,
		pb.CollaboratorStatus_COLLABORATOR_STATUS_REJECTED:
		return ptrBool(false)
	default:
		return nil
	}
}

func mapAddressType(value string) pb.AddressType {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "HOME":
		return pb.AddressType_ADDRESS_TYPE_HOME
	case "WORK":
		return pb.AddressType_ADDRESS_TYPE_BUSINESS
	case "BILLING":
		return pb.AddressType_ADDRESS_TYPE_BILLING
	case "SHIPPING":
		return pb.AddressType_ADDRESS_TYPE_SHIPPING
	case "WAREHOUSE":
		return pb.AddressType_ADDRESS_TYPE_WAREHOUSE
	default:
		return pb.AddressType_ADDRESS_TYPE_BUSINESS
	}
}

func splitContactName(name string) (string, string) {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "Collaborator", "User"
	}
	first := parts[0]
	last := "User"
	if len(parts) > 1 {
		last = strings.Join(parts[1:], " ")
	}
	return first, last
}

func sanitizeUsername(name string) string {
	trimmed := strings.TrimSpace(strings.ToLower(name))
	if trimmed == "" {
		return "collaborator"
	}
	trimmed = strings.ReplaceAll(trimmed, " ", ".")
	return trimmed
}

func firstNonEmpty(values ...*string) string {
	for _, v := range values {
		if v != nil && strings.TrimSpace(*v) != "" {
			return strings.TrimSpace(*v)
		}
	}
	return ""
}

func grpcErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	if st, ok := status.FromError(err); ok {
		return st.Message()
	}
	return err.Error()
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func ptrString(value string) *string {
	v := value
	return &v
}

func ptrBool(value bool) *bool {
	v := value
	return &v
}
