package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"kisanlink-erp/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// S3Service handles file operations with AWS S3
type S3Service struct {
	client     *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
	bucket     string
	region     string
}

// NewS3Service creates a new S3 service instance
func NewS3Service(cfg *config.Config) (*S3Service, error) {
	// Load AWS configuration
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AWS.Region),
		awsconfig.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     cfg.AWS.AccessKeyID,
				SecretAccessKey: cfg.AWS.SecretAccessKey,
			}, nil
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg)

	// Create uploader and downloader
	uploader := manager.NewUploader(client)
	downloader := manager.NewDownloader(client)

	return &S3Service{
		client:     client,
		uploader:   uploader,
		downloader: downloader,
		bucket:     cfg.AWS.S3Bucket,
		region:     cfg.AWS.Region,
	}, nil
}

// UploadFile uploads a file to S3
func (s *S3Service) UploadFile(ctx context.Context, file *multipart.FileHeader, saleID, returnID *string) (string, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Determine folder based on context
	var folder string
	if saleID != nil {
		folder = fmt.Sprintf("sales/%s", *saleID)
	} else if returnID != nil {
		folder = fmt.Sprintf("returns/%s", *returnID)
	} else {
		folder = "misc"
	}

	key := fmt.Sprintf("%s/%s", folder, filename)

	// Upload to S3
	_, err = s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        src,
		ContentType: aws.String(file.Header.Get("Content-Type")),
		Metadata: map[string]string{
			"original-filename": file.Filename,
			"uploaded-at":       time.Now().UTC().Format(time.RFC3339),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Return the S3 URL
	s3URL := fmt.Sprintf("s3://%s/%s", s.bucket, key)
	return s3URL, nil
}

// DownloadFile downloads a file from S3
func (s *S3Service) DownloadFile(ctx context.Context, s3URL string) (io.ReadCloser, string, error) {
	// Extract key from S3 URL
	key := strings.TrimPrefix(s3URL, fmt.Sprintf("s3://%s/", s.bucket))
	if key == s3URL {
		return nil, "", fmt.Errorf("invalid S3 URL format")
	}

	// Get object from S3
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object from S3: %w", err)
	}

	return result.Body, aws.ToString(result.ContentType), nil
}

// DeleteFile deletes a file from S3
func (s *S3Service) DeleteFile(ctx context.Context, s3URL string) error {
	// Extract key from S3 URL
	key := strings.TrimPrefix(s3URL, fmt.Sprintf("s3://%s/", s.bucket))
	if key == s3URL {
		return fmt.Errorf("invalid S3 URL format")
	}

	// Delete object from S3
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}

// GeneratePresignedURL generates a presigned URL for file access
func (s *S3Service) GeneratePresignedURL(ctx context.Context, s3URL string, expiration time.Duration) (string, error) {
	// Extract key from S3 URL
	key := strings.TrimPrefix(s3URL, fmt.Sprintf("s3://%s/", s.bucket))
	if key == s3URL {
		return "", fmt.Errorf("invalid S3 URL format")
	}

	// Create presigned URL
	presignClient := s3.NewPresignClient(s.client)
	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// FileExists checks if a file exists in S3
func (s *S3Service) FileExists(ctx context.Context, s3URL string) (bool, error) {
	// Extract key from S3 URL
	key := strings.TrimPrefix(s3URL, fmt.Sprintf("s3://%s/", s.bucket))
	if key == s3URL {
		return false, fmt.Errorf("invalid S3 URL format")
	}

	// Check if object exists
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "NoSuchKey") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if file exists: %w", err)
	}

	return true, nil
}

// GetFileInfo gets information about a file in S3
func (s *S3Service) GetFileInfo(ctx context.Context, s3URL string) (*FileInfo, error) {
	// Extract key from S3 URL
	key := strings.TrimPrefix(s3URL, fmt.Sprintf("s3://%s/", s.bucket))
	if key == s3URL {
		return nil, fmt.Errorf("invalid S3 URL format")
	}

	// Get object metadata
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileInfo{
		Key:          key,
		Size:         aws.ToInt64(result.ContentLength),
		ContentType:  aws.ToString(result.ContentType),
		LastModified: aws.ToTime(result.LastModified),
		Metadata:     result.Metadata,
	}, nil
}

// FileInfo represents file information from S3
type FileInfo struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
	Metadata     map[string]string
}

// ValidateFileType checks if the file type is allowed
func (s *S3Service) ValidateFileType(filename string) error {
	allowedExtensions := []string{".pdf", ".jpg", ".jpeg", ".png", ".gif", ".doc", ".docx", ".xls", ".xlsx"}
	ext := strings.ToLower(filepath.Ext(filename))

	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			return nil
		}
	}

	return fmt.Errorf("file type %s is not allowed", ext)
}

// GetFileSize returns the size of the uploaded file
func (s *S3Service) GetFileSize(file *multipart.FileHeader) int64 {
	return file.Size
}

// GetContentType returns the content type of the uploaded file
func (s *S3Service) GetContentType(file *multipart.FileHeader) string {
	return file.Header.Get("Content-Type")
}
