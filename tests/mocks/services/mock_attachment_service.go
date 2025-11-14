package services

import (
	"context"
	"mime/multipart"
	"time"

	"kisanlink-erp/internal/database/models"

	"github.com/stretchr/testify/mock"
)

// MockAttachmentService is a mock implementation of AttachmentServiceInterface
type MockAttachmentService struct {
	mock.Mock
}

func (m *MockAttachmentService) UploadAttachment(ctx context.Context, file *multipart.FileHeader, entityType, entityID, uploadedBy string) (*models.AttachmentResponse, error) {
	args := m.Called(ctx, file, entityType, entityID, uploadedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AttachmentResponse), args.Error(1)
}

func (m *MockAttachmentService) GetAttachment(id string) (*models.Attachment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Attachment), args.Error(1)
}

func (m *MockAttachmentService) GetAttachments(entityType, entityID *string, limit, offset int) ([]models.AttachmentResponse, error) {
	args := m.Called(entityType, entityID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AttachmentResponse), args.Error(1)
}

func (m *MockAttachmentService) GetAttachmentsByEntity(entityType, entityID string) ([]models.AttachmentResponse, error) {
	args := m.Called(entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AttachmentResponse), args.Error(1)
}

func (m *MockAttachmentService) DeleteAttachment(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAttachmentService) DownloadAttachment(ctx context.Context, id string) (interface{}, string, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.String(1), args.Error(2)
}

func (m *MockAttachmentService) GenerateDownloadURL(ctx context.Context, id string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, id, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockAttachmentService) GetAttachmentInfo(ctx context.Context, id string) (*models.AttachmentInfoResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AttachmentInfoResponse), args.Error(1)
}
