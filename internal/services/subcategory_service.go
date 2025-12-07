package services

import (
	"context"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
)

// SubcategoryService handles subcategory business logic
type SubcategoryService struct {
	subcategoryRepo *repositories.SubcategoryRepository
	categoryRepo    *repositories.CategoryRepository
	logger          interfaces.Logger
}

// NewSubcategoryService creates a new subcategory service
func NewSubcategoryService(subcategoryRepo *repositories.SubcategoryRepository, categoryRepo *repositories.CategoryRepository, logger interfaces.Logger) *SubcategoryService {
	return &SubcategoryService{
		subcategoryRepo: subcategoryRepo,
		categoryRepo:    categoryRepo,
		logger:          logger,
	}
}

// CreateSubcategory creates a new subcategory
func (s *SubcategoryService) CreateSubcategory(ctx context.Context, request *models.CreateSubcategoryRequest) (*models.SubcategoryResponse, error) {
	s.logger.Info("Creating subcategory",
		zap.String("name", request.Name),
		zap.String("category", request.CategoryName))

	// Validate category exists
	categoryExists, err := s.categoryRepo.ExistsByName(request.CategoryName)
	if err != nil {
		s.logger.Error("Failed to check category existence", zap.Error(err))
		return nil, err
	}
	if !categoryExists {
		return nil, errors.NewBadRequestError("Category does not exist")
	}

	// Check if subcategory already exists
	existing, err := s.subcategoryRepo.GetByName(request.Name)
	if err != nil {
		s.logger.Error("Failed to check subcategory existence", zap.Error(err))
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewBadRequestError("Subcategory with this name already exists")
	}

	subcategory := models.NewSubcategory(request.Name, request.CategoryName, request.Description)

	if err := s.subcategoryRepo.Create(subcategory); err != nil {
		s.logger.Error("Failed to create subcategory", zap.Error(err))
		return nil, err
	}

	response := &models.SubcategoryResponse{
		ID:           subcategory.ID,
		Name:         subcategory.Name,
		Description:  subcategory.Description,
		CategoryName: subcategory.CategoryName,
		CreatedAt:    subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("Subcategory created successfully", zap.String("id", subcategory.ID))
	return response, nil
}

// GetSubcategory retrieves a subcategory by ID
func (s *SubcategoryService) GetSubcategory(ctx context.Context, id string) (*models.SubcategoryResponse, error) {
	s.logger.Info("Retrieving subcategory", zap.String("id", id))

	subcategory, err := s.subcategoryRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve subcategory", zap.Error(err))
		return nil, err
	}

	response := &models.SubcategoryResponse{
		ID:           subcategory.ID,
		Name:         subcategory.Name,
		Description:  subcategory.Description,
		CategoryName: subcategory.CategoryName,
		CreatedAt:    subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response, nil
}

// GetSubcategoryByName retrieves a subcategory by name
func (s *SubcategoryService) GetSubcategoryByName(ctx context.Context, name string) (*models.SubcategoryResponse, error) {
	s.logger.Info("Retrieving subcategory by name", zap.String("name", name))

	subcategory, err := s.subcategoryRepo.GetByName(name)
	if err != nil {
		s.logger.Error("Failed to retrieve subcategory by name", zap.Error(err))
		return nil, err
	}
	if subcategory == nil {
		return nil, errors.NewNotFoundError("Subcategory")
	}

	response := &models.SubcategoryResponse{
		ID:           subcategory.ID,
		Name:         subcategory.Name,
		Description:  subcategory.Description,
		CategoryName: subcategory.CategoryName,
		CreatedAt:    subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response, nil
}

// GetSubcategoriesByCategory retrieves all subcategories for a category
func (s *SubcategoryService) GetSubcategoriesByCategory(ctx context.Context, categoryName string) ([]models.SubcategoryResponse, error) {
	s.logger.Info("Retrieving subcategories by category", zap.String("category", categoryName))

	subcategories, err := s.subcategoryRepo.GetByCategoryName(categoryName)
	if err != nil {
		s.logger.Error("Failed to retrieve subcategories by category", zap.Error(err))
		return nil, err
	}

	var responses []models.SubcategoryResponse
	for _, subcategory := range subcategories {
		response := models.SubcategoryResponse{
			ID:           subcategory.ID,
			Name:         subcategory.Name,
			Description:  subcategory.Description,
			CategoryName: subcategory.CategoryName,
			CreatedAt:    subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:    subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	s.logger.Info("Retrieved subcategories by category",
		zap.String("category", categoryName),
		zap.Int("count", len(responses)))
	return responses, nil
}

// GetAllSubcategories retrieves all subcategories
func (s *SubcategoryService) GetAllSubcategories(ctx context.Context) ([]models.SubcategoryResponse, error) {
	s.logger.Info("Retrieving all subcategories")

	subcategories, err := s.subcategoryRepo.GetAll()
	if err != nil {
		s.logger.Error("Failed to retrieve all subcategories", zap.Error(err))
		return nil, err
	}

	var responses []models.SubcategoryResponse
	for _, subcategory := range subcategories {
		response := models.SubcategoryResponse{
			ID:           subcategory.ID,
			Name:         subcategory.Name,
			Description:  subcategory.Description,
			CategoryName: subcategory.CategoryName,
			CreatedAt:    subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:    subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	s.logger.Info("Retrieved all subcategories", zap.Int("count", len(responses)))
	return responses, nil
}

// UpdateSubcategory updates a subcategory
func (s *SubcategoryService) UpdateSubcategory(ctx context.Context, id string, request *models.UpdateSubcategoryRequest) (*models.SubcategoryResponse, error) {
	s.logger.Info("Updating subcategory", zap.String("id", id))

	subcategory, err := s.subcategoryRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve subcategory for update", zap.Error(err))
		return nil, err
	}

	if request.Name != nil {
		// Check if new name already exists
		existing, err := s.subcategoryRepo.GetByName(*request.Name)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, errors.NewBadRequestError("Subcategory with this name already exists")
		}
		subcategory.Name = *request.Name
	}
	if request.Description != nil {
		subcategory.Description = request.Description
	}

	if err := s.subcategoryRepo.Update(subcategory); err != nil {
		s.logger.Error("Failed to update subcategory", zap.Error(err))
		return nil, err
	}

	response := &models.SubcategoryResponse{
		ID:           subcategory.ID,
		Name:         subcategory.Name,
		Description:  subcategory.Description,
		CategoryName: subcategory.CategoryName,
		CreatedAt:    subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("Subcategory updated successfully", zap.String("id", id))
	return response, nil
}

// DeleteSubcategory deletes a subcategory
func (s *SubcategoryService) DeleteSubcategory(ctx context.Context, id string) error {
	s.logger.Info("Deleting subcategory", zap.String("id", id))

	exists, err := s.subcategoryRepo.Exists(id)
	if err != nil {
		s.logger.Error("Failed to check subcategory existence", zap.Error(err))
		return err
	}
	if !exists {
		return errors.NewNotFoundError("Subcategory")
	}

	if err := s.subcategoryRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete subcategory", zap.Error(err))
		return err
	}

	s.logger.Info("Subcategory deleted successfully", zap.String("id", id))
	return nil
}
