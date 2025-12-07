package services

import (
	"context"

	"kisanlink-erp/internal/database/models"
	"kisanlink-erp/internal/database/repositories"
	"kisanlink-erp/internal/errors"
	"kisanlink-erp/internal/interfaces"

	"go.uber.org/zap"
)

// CategoryService handles category business logic
type CategoryService struct {
	categoryRepo    *repositories.CategoryRepository
	subcategoryRepo *repositories.SubcategoryRepository
	logger          interfaces.Logger
}

// NewCategoryService creates a new category service
func NewCategoryService(categoryRepo *repositories.CategoryRepository, subcategoryRepo *repositories.SubcategoryRepository, logger interfaces.Logger) *CategoryService {
	return &CategoryService{
		categoryRepo:    categoryRepo,
		subcategoryRepo: subcategoryRepo,
		logger:          logger,
	}
}

// predefinedCategories contains the default category and subcategory hierarchy
var predefinedCategories = []struct {
	Name          string
	Subcategories []string
}{
	{"Seeds", nil},
	{"Fertilizers", []string{"BULK", "Water Soluble", "Micronutrients", "Macronutrients"}},
	{"Pesticides", []string{"Weedicides", "Insecticides", "Fungicides", "Organic"}},
	{"Bio Products", []string{"Bulk", "Liquids", "Others"}},
	{"Implements", []string{"Weeding", "Sowing", "Sprayers"}},
	{"Irrigation", []string{"Pipes", "Drippers", "Sprinklers", "Automation Machines", "Others"}},
	{"Others", nil},
}

// SeedCategories seeds all predefined categories and subcategories (idempotent)
func (s *CategoryService) SeedCategories(ctx context.Context) (*models.SeedCategoriesResponse, error) {
	s.logger.Info("Starting category seeding")

	categoriesCreated := 0
	subcategoriesCreated := 0

	for _, cat := range predefinedCategories {
		// Create or get category
		category := models.NewCategory(cat.Name, nil)
		existingCat, err := s.categoryRepo.GetByName(cat.Name)
		if err != nil {
			s.logger.Error("Failed to check category existence",
				zap.Error(err),
				zap.String("name", cat.Name))
			return nil, err
		}

		if existingCat == nil {
			if err := s.categoryRepo.Create(category); err != nil {
				s.logger.Error("Failed to create category",
					zap.Error(err),
					zap.String("name", cat.Name))
				return nil, err
			}
			categoriesCreated++
			s.logger.Info("Created category", zap.String("name", cat.Name))
		} else {
			s.logger.Debug("Category already exists", zap.String("name", cat.Name))
		}

		// Create subcategories
		for _, subName := range cat.Subcategories {
			existingSub, err := s.subcategoryRepo.GetByName(subName)
			if err != nil {
				s.logger.Error("Failed to check subcategory existence",
					zap.Error(err),
					zap.String("name", subName))
				return nil, err
			}

			if existingSub == nil {
				subcategory := models.NewSubcategory(subName, cat.Name, nil)
				if err := s.subcategoryRepo.Create(subcategory); err != nil {
					s.logger.Error("Failed to create subcategory",
						zap.Error(err),
						zap.String("name", subName),
						zap.String("category", cat.Name))
					return nil, err
				}
				subcategoriesCreated++
				s.logger.Info("Created subcategory",
					zap.String("name", subName),
					zap.String("category", cat.Name))
			} else {
				s.logger.Debug("Subcategory already exists",
					zap.String("name", subName))
			}
		}
	}

	response := &models.SeedCategoriesResponse{
		CategoriesCreated:    categoriesCreated,
		SubcategoriesCreated: subcategoriesCreated,
		Message:              "Categories seeded successfully",
	}

	s.logger.Info("Category seeding completed",
		zap.Int("categories_created", categoriesCreated),
		zap.Int("subcategories_created", subcategoriesCreated))

	return response, nil
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(ctx context.Context, request *models.CreateCategoryRequest) (*models.CategoryResponse, error) {
	s.logger.Info("Creating category", zap.String("name", request.Name))

	// Check if category already exists
	existing, err := s.categoryRepo.GetByName(request.Name)
	if err != nil {
		s.logger.Error("Failed to check category existence", zap.Error(err))
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewBadRequestError("Category with this name already exists")
	}

	category := models.NewCategory(request.Name, request.Description)

	if err := s.categoryRepo.Create(category); err != nil {
		s.logger.Error("Failed to create category", zap.Error(err))
		return nil, err
	}

	response := &models.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   category.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("Category created successfully", zap.String("id", category.ID))
	return response, nil
}

// GetCategory retrieves a category by ID
func (s *CategoryService) GetCategory(ctx context.Context, id string) (*models.CategoryResponse, error) {
	s.logger.Info("Retrieving category", zap.String("id", id))

	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve category", zap.Error(err))
		return nil, err
	}

	response := &models.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   category.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response, nil
}

// GetCategoryByName retrieves a category by name
func (s *CategoryService) GetCategoryByName(ctx context.Context, name string) (*models.CategoryResponse, error) {
	s.logger.Info("Retrieving category by name", zap.String("name", name))

	category, err := s.categoryRepo.GetByName(name)
	if err != nil {
		s.logger.Error("Failed to retrieve category by name", zap.Error(err))
		return nil, err
	}
	if category == nil {
		return nil, errors.NewNotFoundError("Category")
	}

	response := &models.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   category.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response, nil
}

// GetAllCategories retrieves all categories
func (s *CategoryService) GetAllCategories(ctx context.Context) ([]models.CategoryResponse, error) {
	s.logger.Info("Retrieving all categories")

	categories, err := s.categoryRepo.GetAll()
	if err != nil {
		s.logger.Error("Failed to retrieve all categories", zap.Error(err))
		return nil, err
	}

	var responses []models.CategoryResponse
	for _, category := range categories {
		response := models.CategoryResponse{
			ID:          category.ID,
			Name:        category.Name,
			Description: category.Description,
			CreatedAt:   category.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	s.logger.Info("Retrieved all categories", zap.Int("count", len(responses)))
	return responses, nil
}

// GetAllCategoriesWithSubcategories retrieves all categories with their subcategories
func (s *CategoryService) GetAllCategoriesWithSubcategories(ctx context.Context) ([]models.CategoryResponse, error) {
	s.logger.Info("Retrieving all categories with subcategories")

	categories, err := s.categoryRepo.GetAllWithSubcategories()
	if err != nil {
		s.logger.Error("Failed to retrieve categories with subcategories", zap.Error(err))
		return nil, err
	}

	var responses []models.CategoryResponse
	for _, category := range categories {
		var subcatResponses []models.SubcategoryResponse
		for _, subcat := range category.Subcategories {
			subcatResponse := models.SubcategoryResponse{
				ID:           subcat.ID,
				Name:         subcat.Name,
				Description:  subcat.Description,
				CategoryName: subcat.CategoryName,
				CreatedAt:    subcat.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:    subcat.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			}
			subcatResponses = append(subcatResponses, subcatResponse)
		}

		response := models.CategoryResponse{
			ID:            category.ID,
			Name:          category.Name,
			Description:   category.Description,
			Subcategories: subcatResponses,
			CreatedAt:     category.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	s.logger.Info("Retrieved all categories with subcategories", zap.Int("count", len(responses)))
	return responses, nil
}

// UpdateCategory updates a category
func (s *CategoryService) UpdateCategory(ctx context.Context, id string, request *models.UpdateCategoryRequest) (*models.CategoryResponse, error) {
	s.logger.Info("Updating category", zap.String("id", id))

	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve category for update", zap.Error(err))
		return nil, err
	}

	if request.Name != nil {
		// Check if new name already exists
		existing, err := s.categoryRepo.GetByName(*request.Name)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, errors.NewBadRequestError("Category with this name already exists")
		}
		category.Name = *request.Name
	}
	if request.Description != nil {
		category.Description = request.Description
	}

	if err := s.categoryRepo.Update(category); err != nil {
		s.logger.Error("Failed to update category", zap.Error(err))
		return nil, err
	}

	response := &models.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   category.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("Category updated successfully", zap.String("id", id))
	return response, nil
}

// DeleteCategory deletes a category
func (s *CategoryService) DeleteCategory(ctx context.Context, id string) error {
	s.logger.Info("Deleting category", zap.String("id", id))

	exists, err := s.categoryRepo.Exists(id)
	if err != nil {
		s.logger.Error("Failed to check category existence", zap.Error(err))
		return err
	}
	if !exists {
		return errors.NewNotFoundError("Category")
	}

	if err := s.categoryRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete category", zap.Error(err))
		return err
	}

	s.logger.Info("Category deleted successfully", zap.String("id", id))
	return nil
}
