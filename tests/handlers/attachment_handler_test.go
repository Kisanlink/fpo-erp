package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"kisanlink-erp/internal/api/handlers"
	"kisanlink-erp/internal/database/models"
	mockServices "kisanlink-erp/tests/mocks/services"
	"kisanlink-erp/tests/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	// Bypass AAA for testing
	os.Setenv("AAA_ENABLED", "false")
	gin.SetMode(gin.TestMode)
}

// TestAttachmentHandler_UploadAttachment_Success tests successful attachment upload
func TestAttachmentHandler_UploadAttachment_Success(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	expectedResponse := &models.AttachmentResponse{
		ID:         "ATCH00000001",
		EntityType: "logo",
		EntityID:   "CLAB00000001",
		FilePath:   "logos/test.png",
		FileType:   "image/png",
		UploadedBy: stringPtr("USER_12345678"),
		UploadedAt: "2025-11-10T10:00:00Z",
	}

	mockService.On("UploadAttachment", mock.Anything, mock.Anything, "logo", "CLAB00000001", "test-user-123").Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("entity_type", "logo")
	writer.WriteField("entity_id", "CLAB00000001")
	part, _ := writer.CreateFormFile("file", "test.png")
	part.Write([]byte("test file content"))
	writer.Close()

	c.Request, _ = http.NewRequest("POST", "/api/v1/attachments", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Set("user_id", "test-user-123")

	handler.UploadAttachment(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Attachment uploaded successfully", response["message"])
	assert.NotNil(t, response["data"])
}

// TestAttachmentHandler_UploadAttachment_MissingEntityType tests missing entity_type
func TestAttachmentHandler_UploadAttachment_MissingEntityType(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create multipart form without entity_type
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("entity_id", "CLAB00000001")
	part, _ := writer.CreateFormFile("file", "test.png")
	part.Write([]byte("test file content"))
	writer.Close()

	c.Request, _ = http.NewRequest("POST", "/api/v1/attachments", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Set("user_id", "test-user-123")

	handler.UploadAttachment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "UploadAttachment", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// TestAttachmentHandler_UploadAttachment_ServiceError tests service layer errors
func TestAttachmentHandler_UploadAttachment_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	mockService.On("UploadAttachment", mock.Anything, mock.Anything, "logo", "CLAB00000001", "test-user-123").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("entity_type", "logo")
	writer.WriteField("entity_id", "CLAB00000001")
	part, _ := writer.CreateFormFile("file", "test.png")
	part.Write([]byte("test file content"))
	writer.Close()

	c.Request, _ = http.NewRequest("POST", "/api/v1/attachments", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Set("user_id", "test-user-123")

	handler.UploadAttachment(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestAttachmentHandler_GetAttachment_Success tests fetching single attachment
func TestAttachmentHandler_GetAttachment_Success(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	uploadedBy := "USER_12345678"
	expectedAttachment := &models.Attachment{
		EntityType: "logo",
		EntityID:   "CLAB00000001",
		FilePath:   "logos/test.png",
		FileType:   "image/png",
		UploadedBy: &uploadedBy,
	}

	mockService.On("GetAttachment", "ATCH00000001").Return(expectedAttachment, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "ATCH00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments/ATCH00000001", nil)

	handler.GetAttachment(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Attachment retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
}

// TestAttachmentHandler_GetAttachment_ServiceError tests service error
func TestAttachmentHandler_GetAttachment_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	mockService.On("GetAttachment", "ATCH99999999").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "ATCH99999999"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments/ATCH99999999", nil)

	handler.GetAttachment(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestAttachmentHandler_GetAttachments_Success tests fetching all attachments
func TestAttachmentHandler_GetAttachments_Success(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	expectedAttachments := []models.AttachmentResponse{
		{
			ID:         "ATCH00000001",
			EntityType: "logo",
			EntityID:   "CLAB00000001",
			FilePath:   "logos/test.png",
			FileType:   "image/png",
		},
		{
			ID:         "ATCH00000002",
			EntityType: "po",
			EntityID:   "PORD00000001",
			FilePath:   "purchase-orders/PO_xxx/doc.pdf",
			FileType:   "application/pdf",
		},
	}

	mockService.On("GetAttachments", (*string)(nil), (*string)(nil), 10, 0).Return(expectedAttachments, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments", nil)

	handler.GetAttachments(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Attachments retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
}

// TestAttachmentHandler_GetAttachments_WithFilters tests with entity filters
func TestAttachmentHandler_GetAttachments_WithFilters(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	mockService.On("GetAttachments", stringPtr("logo"), stringPtr("CLAB00000001"), 10, 0).Return([]models.AttachmentResponse{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments?entity_type=logo&entity_id=CLAB00000001", nil)

	handler.GetAttachments(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestAttachmentHandler_GetAttachments_WithPagination tests pagination
func TestAttachmentHandler_GetAttachments_WithPagination(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	mockService.On("GetAttachments", (*string)(nil), (*string)(nil), 20, 10).Return([]models.AttachmentResponse{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments?limit=20&offset=10", nil)

	handler.GetAttachments(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestAttachmentHandler_DownloadAttachment_Success tests file download
func TestAttachmentHandler_DownloadAttachment_Success(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	fileContent := strings.NewReader("test file content")
	mockService.On("DownloadAttachment", mock.Anything, "ATCH00000001").Return(fileContent, "image/png", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "ATCH00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments/ATCH00000001/download", nil)

	handler.DownloadAttachment(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	mockService.AssertExpectations(t)
}

// TestAttachmentHandler_DownloadAttachment_ServiceError tests download service error
func TestAttachmentHandler_DownloadAttachment_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	mockService.On("DownloadAttachment", mock.Anything, "ATCH99999999").Return(nil, "", assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "ATCH99999999"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments/ATCH99999999/download", nil)

	handler.DownloadAttachment(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestAttachmentHandler_GenerateDownloadURL_Success tests presigned URL generation
func TestAttachmentHandler_GenerateDownloadURL_Success(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	expectedURL := "https://s3.amazonaws.com/bucket/logos/test.png?presigned=true"
	mockService.On("GenerateDownloadURL", mock.Anything, "ATCH00000001", time.Duration(3600)*time.Second).Return(expectedURL, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "ATCH00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments/ATCH00000001/url", nil)

	handler.GenerateDownloadURL(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Download URL generated successfully", response["message"])
	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["download_url"])
	assert.Equal(t, float64(3600), data["expires_in"])
}

// TestAttachmentHandler_GenerateDownloadURL_CustomExpiration tests custom expiration
func TestAttachmentHandler_GenerateDownloadURL_CustomExpiration(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	expectedURL := "https://s3.amazonaws.com/bucket/logos/test.png?presigned=true"
	mockService.On("GenerateDownloadURL", mock.Anything, "ATCH00000001", time.Duration(7200)*time.Second).Return(expectedURL, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "ATCH00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments/ATCH00000001/url?expiration=7200", nil)

	handler.GenerateDownloadURL(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(7200), data["expires_in"])
}

// TestAttachmentHandler_GenerateDownloadURL_ServiceError tests URL generation error
func TestAttachmentHandler_GenerateDownloadURL_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	mockService.On("GenerateDownloadURL", mock.Anything, "ATCH99999999", mock.Anything).Return("", assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "ATCH99999999"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments/ATCH99999999/url", nil)

	handler.GenerateDownloadURL(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestAttachmentHandler_GetAttachmentInfo_Success tests getting detailed attachment info
func TestAttachmentHandler_GetAttachmentInfo_Success(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	expectedInfo := &models.AttachmentInfoResponse{
		ID:         "ATCH00000001",
		EntityType: "logo",
		EntityID:   "CLAB00000001",
		FilePath:   "logos/test.png",
		FileType:   "image/png",
		FileSize:   1024,
		UploadedBy: stringPtr("USER_12345678"),
		UploadedAt: "2025-11-10T10:00:00Z",
	}

	mockService.On("GetAttachmentInfo", mock.Anything, "ATCH00000001").Return(expectedInfo, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "ATCH00000001"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments/ATCH00000001/info", nil)

	handler.GetAttachmentInfo(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Attachment info retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
}

// TestAttachmentHandler_GetAttachmentInfo_ServiceError tests info retrieval error
func TestAttachmentHandler_GetAttachmentInfo_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	mockService.On("GetAttachmentInfo", mock.Anything, "ATCH99999999").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "ATCH99999999"}}
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments/ATCH99999999/info", nil)

	handler.GetAttachmentInfo(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestAttachmentHandler_GetAttachmentsByEntity_Success tests entity-based retrieval
func TestAttachmentHandler_GetAttachmentsByEntity_Success(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	expectedAttachments := []models.AttachmentResponse{
		{
			ID:         "ATCH00000001",
			EntityType: "logo",
			EntityID:   "CLAB00000001",
			FilePath:   "logos/test.png",
			FileType:   "image/png",
		},
	}

	mockService.On("GetAttachmentsByEntity", "logo", "CLAB00000001").Return(expectedAttachments, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		{Key: "entity_type", Value: "logo"},
		{Key: "entity_id", Value: "CLAB00000001"},
	}
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments/entity/logo/CLAB00000001", nil)

	handler.GetAttachmentsByEntity(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Entity attachments retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
}

// TestAttachmentHandler_GetAttachmentsByEntity_ServiceError tests entity retrieval error
func TestAttachmentHandler_GetAttachmentsByEntity_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	mockService.On("GetAttachmentsByEntity", "logo", "CLAB99999999").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		{Key: "entity_type", Value: "logo"},
		{Key: "entity_id", Value: "CLAB99999999"},
	}
	c.Request, _ = http.NewRequest("GET", "/api/v1/attachments/entity/logo/CLAB99999999", nil)

	handler.GetAttachmentsByEntity(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// TestAttachmentHandler_DeleteAttachment_Success tests attachment deletion
func TestAttachmentHandler_DeleteAttachment_Success(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	mockService.On("DeleteAttachment", mock.Anything, "ATCH00000001").Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "ATCH00000001"}}
	c.Request, _ = http.NewRequest("DELETE", "/api/v1/attachments/ATCH00000001", nil)

	handler.DeleteAttachment(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Attachment deleted successfully", response["message"])
}

// TestAttachmentHandler_DeleteAttachment_ServiceError tests deletion error
func TestAttachmentHandler_DeleteAttachment_ServiceError(t *testing.T) {
	mockService := new(mockServices.MockAttachmentService)
	mockAAA := testutils.NewMockAAAMiddleware()
	handler := handlers.NewAttachmentHandler(mockService, mockAAA)

	mockService.On("DeleteAttachment", mock.Anything, "ATCH99999999").Return(assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "ATCH99999999"}}
	c.Request, _ = http.NewRequest("DELETE", "/api/v1/attachments/ATCH99999999", nil)

	handler.DeleteAttachment(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
