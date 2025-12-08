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
// Name is automatically normalized to UPPER_SNAKE_CASE (e.g., "water soluble" -> "WATER_SOLUBLE")
func (s *SubcategoryService) CreateSubcategory(ctx context.Context, request *models.CreateSubcategoryRequest) (*models.SubcategoryResponse, error) {
	// Normalize name to UPPER_SNAKE_CASE
	normalizedName := toSnakeCase(request.Name)
	s.logger.Info("Creating subcategory",
		zap.String("original_name", request.Name),
		zap.String("normalized_name", normalizedName),
		zap.String("category_id", request.CategoryID))

	// Validate category exists by ID
	categoryExists, err := s.categoryRepo.Exists(request.CategoryID)
	if err != nil {
		s.logger.Error("Failed to check category existence", zap.Error(err))
		return nil, err
	}
	if !categoryExists {
		return nil, errors.NewBadRequestError("Category does not exist")
	}

	// Check if subcategory already exists in this category (name is unique per category)
	existing, err := s.subcategoryRepo.GetByNameAndCategoryID(normalizedName, request.CategoryID)
	if err != nil {
		s.logger.Error("Failed to check subcategory existence", zap.Error(err))
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewBadRequestError("Subcategory with this name already exists in this category")
	}

	subcategory := models.NewSubcategory(normalizedName, request.CategoryID, request.Description)

	if err := s.subcategoryRepo.Create(subcategory); err != nil {
		s.logger.Error("Failed to create subcategory", zap.Error(err))
		return nil, err
	}

	response := &models.SubcategoryResponse{
		ID:          subcategory.ID,
		Name:        subcategory.Name,
		Description: subcategory.Description,
		CategoryID:  subcategory.CategoryID,
		CreatedAt:   subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
		ID:          subcategory.ID,
		Name:        subcategory.Name,
		Description: subcategory.Description,
		CategoryID:  subcategory.CategoryID,
		CreatedAt:   subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response, nil
}

// GetSubcategoryByName retrieves a subcategory by name
// Note: Name is not globally unique - returns first match
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
		ID:          subcategory.ID,
		Name:        subcategory.Name,
		Description: subcategory.Description,
		CategoryID:  subcategory.CategoryID,
		CreatedAt:   subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response, nil
}

// GetSubcategoriesByCategory retrieves all subcategories for a category by ID
func (s *SubcategoryService) GetSubcategoriesByCategory(ctx context.Context, categoryID string) ([]models.SubcategoryResponse, error) {
	s.logger.Info("Retrieving subcategories by category", zap.String("category_id", categoryID))

	subcategories, err := s.subcategoryRepo.GetByCategoryID(categoryID)
	if err != nil {
		s.logger.Error("Failed to retrieve subcategories by category", zap.Error(err))
		return nil, err
	}

	var responses []models.SubcategoryResponse
	for _, subcategory := range subcategories {
		response := models.SubcategoryResponse{
			ID:          subcategory.ID,
			Name:        subcategory.Name,
			Description: subcategory.Description,
			CategoryID:  subcategory.CategoryID,
			CreatedAt:   subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	s.logger.Info("Retrieved subcategories by category",
		zap.String("category_id", categoryID),
		zap.Int("count", len(responses)))
	return responses, nil
}

// GetAllSubcategories retrieves all subcategories (non-paginated, for internal use)
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
			ID:          subcategory.ID,
			Name:        subcategory.Name,
			Description: subcategory.Description,
			CategoryID:  subcategory.CategoryID,
			CreatedAt:   subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	s.logger.Info("Retrieved all subcategories", zap.Int("count", len(responses)))
	return responses, nil
}

// GetAllSubcategoriesPaginated retrieves all subcategories with pagination
func (s *SubcategoryService) GetAllSubcategoriesPaginated(ctx context.Context, limit, offset int) ([]models.SubcategoryResponse, int64, error) {
	s.logger.Info("Retrieving all subcategories with pagination",
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	subcategories, total, err := s.subcategoryRepo.GetAllPaginated(limit, offset)
	if err != nil {
		s.logger.Error("Failed to retrieve paginated subcategories", zap.Error(err))
		return nil, 0, err
	}

	var responses []models.SubcategoryResponse
	for _, subcategory := range subcategories {
		response := models.SubcategoryResponse{
			ID:          subcategory.ID,
			Name:        subcategory.Name,
			Description: subcategory.Description,
			CategoryID:  subcategory.CategoryID,
			CreatedAt:   subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	s.logger.Info("Retrieved paginated subcategories",
		zap.Int("count", len(responses)),
		zap.Int64("total", total))
	return responses, total, nil
}

// GetSubcategoriesByCategoryPaginated retrieves subcategories by category with pagination
func (s *SubcategoryService) GetSubcategoriesByCategoryPaginated(ctx context.Context, categoryID string, limit, offset int) ([]models.SubcategoryResponse, int64, error) {
	s.logger.Info("Retrieving subcategories by category with pagination",
		zap.String("category_id", categoryID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	subcategories, total, err := s.subcategoryRepo.GetByCategoryIDPaginated(categoryID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to retrieve paginated subcategories by category", zap.Error(err))
		return nil, 0, err
	}

	var responses []models.SubcategoryResponse
	for _, subcategory := range subcategories {
		response := models.SubcategoryResponse{
			ID:          subcategory.ID,
			Name:        subcategory.Name,
			Description: subcategory.Description,
			CategoryID:  subcategory.CategoryID,
			CreatedAt:   subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	s.logger.Info("Retrieved paginated subcategories by category",
		zap.String("category_id", categoryID),
		zap.Int("count", len(responses)),
		zap.Int64("total", total))
	return responses, total, nil
}

// SearchSubcategories searches subcategories by name with pagination
func (s *SubcategoryService) SearchSubcategories(ctx context.Context, query string, limit, offset int) ([]models.SubcategoryResponse, int64, error) {
	s.logger.Info("Searching subcategories",
		zap.String("query", query),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	subcategories, total, err := s.subcategoryRepo.SearchByNamePaginated(query, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search subcategories", zap.Error(err))
		return nil, 0, err
	}

	var responses []models.SubcategoryResponse
	for _, subcategory := range subcategories {
		response := models.SubcategoryResponse{
			ID:          subcategory.ID,
			Name:        subcategory.Name,
			Description: subcategory.Description,
			CategoryID:  subcategory.CategoryID,
			CreatedAt:   subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	s.logger.Info("Search subcategories completed",
		zap.String("query", query),
		zap.Int("count", len(responses)),
		zap.Int64("total", total))
	return responses, total, nil
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
		// Check if new name already exists in the same category
		existing, err := s.subcategoryRepo.GetByNameAndCategoryID(*request.Name, subcategory.CategoryID)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, errors.NewBadRequestError("Subcategory with this name already exists in this category")
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
		ID:          subcategory.ID,
		Name:        subcategory.Name,
		Description: subcategory.Description,
		CategoryID:  subcategory.CategoryID,
		CreatedAt:   subcategory.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   subcategory.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
