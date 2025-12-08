package services

import (
	"context"
	"strings"

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

// toSnakeCase converts a string to UPPER_SNAKE_CASE format
// Examples: "water soluble" -> "WATER_SOLUBLE", "Bio Products" -> "BIO_PRODUCTS"
func toSnakeCase(s string) string {
	// Trim spaces
	s = strings.TrimSpace(s)
	// Replace spaces with underscores
	s = strings.ReplaceAll(s, " ", "_")
	// Replace hyphens with underscores
	s = strings.ReplaceAll(s, "-", "_")
	// Convert to uppercase
	return strings.ToUpper(s)
}

// SubcategoryDef defines a subcategory with name and description
type SubcategoryDef struct {
	Name        string
	Description string
}

// CategoryDef defines a category with its subcategories
type CategoryDef struct {
	Name          string
	Description   string
	Subcategories []SubcategoryDef
}

// PredefinedCategories contains the default category and subcategory hierarchy (EXPORTED)
// Names are in ALL_CAPS_SNAKE_CASE to indicate they are enumerations
// Used by both CategoryService.SeedCategories() and database migration
var PredefinedCategories = []CategoryDef{
	{"SEEDS", "Agricultural seeds and planting materials", nil},
	{"FERTILIZERS", "Chemical and organic fertilizers", []SubcategoryDef{
		{"BULK", "Bulk fertilizers"},
		{"WATER_SOLUBLE", "Water soluble fertilizers"},
		{"MICRONUTRIENTS", "Micronutrient fertilizers"},
		{"MACRONUTRIENTS", "Macronutrient fertilizers"},
	}},
	{"PESTICIDES", "Pest control products", []SubcategoryDef{
		{"WEEDICIDES", "Weed control products"},
		{"INSECTICIDES", "Insect control products"},
		{"FUNGICIDES", "Fungus control products"},
		{"ORGANIC", "Organic pest control"},
	}},
	{"BIO_PRODUCTS", "Biological and eco-friendly products", []SubcategoryDef{
		{"BULK", "Bulk bio products"},
		{"LIQUIDS", "Liquid bio products"},
		{"OTHER", "Other bio products"},
	}},
	{"IMPLEMENTS", "Agricultural tools and implements", []SubcategoryDef{
		{"WEEDING", "Weeding tools"},
		{"SOWING", "Sowing implements"},
		{"SPRAYERS", "Sprayer equipment"},
	}},
	{"IRRIGATION", "Irrigation equipment and systems", []SubcategoryDef{
		{"PIPES", "Irrigation pipes"},
		{"DRIPPERS", "Drip irrigation"},
		{"SPRINKLERS", "Sprinkler systems"},
		{"AUTOMATION_MACHINES", "Automated irrigation"},
		{"OTHER", "Other irrigation equipment"},
	}},
	{"OTHER", "Miscellaneous agricultural products", nil},
}

// SeedCategories seeds all predefined categories and subcategories (idempotent)
// Uses case-insensitive checks for idempotency
func (s *CategoryService) SeedCategories(ctx context.Context) (*models.SeedCategoriesResponse, error) {
	s.logger.Info("Starting category seeding")

	categoriesCreated := 0
	subcategoriesCreated := 0

	for _, cat := range PredefinedCategories {
		// Check if category already exists (case-insensitive)
		existingCat, err := s.categoryRepo.GetByNameCaseInsensitive(cat.Name)
		if err != nil {
			s.logger.Error("Failed to check category existence",
				zap.Error(err),
				zap.String("name", cat.Name))
			return nil, err
		}

		var categoryID string
		if existingCat == nil {
			// Create category with description
			description := cat.Description
			category := models.NewCategory(cat.Name, &description)
			if err := s.categoryRepo.Create(category); err != nil {
				s.logger.Error("Failed to create category",
					zap.Error(err),
					zap.String("name", cat.Name))
				return nil, err
			}
			categoryID = category.ID
			categoriesCreated++
			s.logger.Info("Created category", zap.String("name", cat.Name))
		} else {
			categoryID = existingCat.ID
			s.logger.Debug("Category already exists", zap.String("name", cat.Name))
		}

		// Create subcategories using category ID (not category name)
		for _, sub := range cat.Subcategories {
			// Check if subcategory exists (case-insensitive name + category_id)
			existingSub, err := s.subcategoryRepo.GetByNameAndCategoryID(sub.Name, categoryID)
			if err != nil {
				s.logger.Error("Failed to check subcategory existence",
					zap.Error(err),
					zap.String("name", sub.Name))
				return nil, err
			}

			if existingSub == nil {
				// Create subcategory with description using category ID
				description := sub.Description
				subcategory := models.NewSubcategory(sub.Name, categoryID, &description)
				if err := s.subcategoryRepo.Create(subcategory); err != nil {
					s.logger.Error("Failed to create subcategory",
						zap.Error(err),
						zap.String("name", sub.Name),
						zap.String("category_id", categoryID))
					return nil, err
				}
				subcategoriesCreated++
				s.logger.Info("Created subcategory",
					zap.String("name", sub.Name),
					zap.String("category_id", categoryID))
			} else {
				s.logger.Debug("Subcategory already exists",
					zap.String("name", sub.Name))
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
// Name is automatically normalized to UPPER_SNAKE_CASE (e.g., "water soluble" -> "WATER_SOLUBLE")
func (s *CategoryService) CreateCategory(ctx context.Context, request *models.CreateCategoryRequest) (*models.CategoryResponse, error) {
	// Normalize name to UPPER_SNAKE_CASE
	normalizedName := toSnakeCase(request.Name)
	s.logger.Info("Creating category",
		zap.String("original_name", request.Name),
		zap.String("normalized_name", normalizedName))

	// Check if category already exists (using normalized name)
	existing, err := s.categoryRepo.GetByName(normalizedName)
	if err != nil {
		s.logger.Error("Failed to check category existence", zap.Error(err))
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewBadRequestError("Category with this name already exists")
	}

	category := models.NewCategory(normalizedName, request.Description)

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

// GetCategory retrieves a category by ID with its subcategories
func (s *CategoryService) GetCategory(ctx context.Context, id string) (*models.CategoryResponse, error) {
	s.logger.Info("Retrieving category with subcategories", zap.String("id", id))

	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to retrieve category", zap.Error(err))
		return nil, err
	}

	// Fetch subcategories for this category
	subcategories, err := s.subcategoryRepo.GetByCategoryID(category.ID)
	if err != nil {
		s.logger.Error("Failed to retrieve subcategories for category",
			zap.Error(err),
			zap.String("category_id", category.ID))
		return nil, err
	}

	var subcatResponses []models.SubcategoryResponse
	for _, subcat := range subcategories {
		subcatResponse := models.SubcategoryResponse{
			ID:          subcat.ID,
			Name:        subcat.Name,
			Description: subcat.Description,
			CategoryID:  subcat.CategoryID,
			CreatedAt:   subcat.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   subcat.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		subcatResponses = append(subcatResponses, subcatResponse)
	}

	response := &models.CategoryResponse{
		ID:            category.ID,
		Name:          category.Name,
		Description:   category.Description,
		Subcategories: subcatResponses,
		CreatedAt:     category.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response, nil
}

// GetCategoryByName retrieves a category by name with its subcategories
func (s *CategoryService) GetCategoryByName(ctx context.Context, name string) (*models.CategoryResponse, error) {
	s.logger.Info("Retrieving category by name with subcategories", zap.String("name", name))

	category, err := s.categoryRepo.GetByName(name)
	if err != nil {
		s.logger.Error("Failed to retrieve category by name", zap.Error(err))
		return nil, err
	}
	if category == nil {
		return nil, errors.NewNotFoundError("Category")
	}

	// Fetch subcategories for this category
	subcategories, err := s.subcategoryRepo.GetByCategoryID(category.ID)
	if err != nil {
		s.logger.Error("Failed to retrieve subcategories for category",
			zap.Error(err),
			zap.String("category_id", category.ID))
		return nil, err
	}

	var subcatResponses []models.SubcategoryResponse
	for _, subcat := range subcategories {
		subcatResponse := models.SubcategoryResponse{
			ID:          subcat.ID,
			Name:        subcat.Name,
			Description: subcat.Description,
			CategoryID:  subcat.CategoryID,
			CreatedAt:   subcat.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   subcat.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		subcatResponses = append(subcatResponses, subcatResponse)
	}

	response := &models.CategoryResponse{
		ID:            category.ID,
		Name:          category.Name,
		Description:   category.Description,
		Subcategories: subcatResponses,
		CreatedAt:     category.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response, nil
}

// GetAllCategories retrieves all categories with their subcategories (non-paginated, for internal use)
func (s *CategoryService) GetAllCategories(ctx context.Context) ([]models.CategoryResponse, error) {
	s.logger.Info("Retrieving all categories with subcategories")

	categories, err := s.categoryRepo.GetAll()
	if err != nil {
		s.logger.Error("Failed to retrieve all categories", zap.Error(err))
		return nil, err
	}

	var responses []models.CategoryResponse
	for _, category := range categories {
		// Fetch subcategories for this category
		subcategories, err := s.subcategoryRepo.GetByCategoryID(category.ID)
		if err != nil {
			s.logger.Error("Failed to retrieve subcategories for category",
				zap.Error(err),
				zap.String("category_id", category.ID))
			return nil, err
		}

		var subcatResponses []models.SubcategoryResponse
		for _, subcat := range subcategories {
			subcatResponse := models.SubcategoryResponse{
				ID:          subcat.ID,
				Name:        subcat.Name,
				Description: subcat.Description,
				CategoryID:  subcat.CategoryID,
				CreatedAt:   subcat.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:   subcat.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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

// GetAllCategoriesPaginated retrieves all categories with pagination
func (s *CategoryService) GetAllCategoriesPaginated(ctx context.Context, limit, offset int) ([]models.CategoryResponse, int64, error) {
	s.logger.Info("Retrieving all categories with pagination",
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	categories, total, err := s.categoryRepo.GetAllPaginated(limit, offset)
	if err != nil {
		s.logger.Error("Failed to retrieve paginated categories", zap.Error(err))
		return nil, 0, err
	}

	var responses []models.CategoryResponse
	for _, category := range categories {
		// Fetch subcategories for this category
		subcategories, err := s.subcategoryRepo.GetByCategoryID(category.ID)
		if err != nil {
			s.logger.Error("Failed to retrieve subcategories for category",
				zap.Error(err),
				zap.String("category_id", category.ID))
			return nil, 0, err
		}

		var subcatResponses []models.SubcategoryResponse
		for _, subcat := range subcategories {
			subcatResponse := models.SubcategoryResponse{
				ID:          subcat.ID,
				Name:        subcat.Name,
				Description: subcat.Description,
				CategoryID:  subcat.CategoryID,
				CreatedAt:   subcat.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:   subcat.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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

	s.logger.Info("Retrieved paginated categories",
		zap.Int("count", len(responses)),
		zap.Int64("total", total))
	return responses, total, nil
}

// SearchCategories searches categories by name with pagination
func (s *CategoryService) SearchCategories(ctx context.Context, query string, limit, offset int) ([]models.CategoryResponse, int64, error) {
	s.logger.Info("Searching categories",
		zap.String("query", query),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	categories, total, err := s.categoryRepo.SearchByNamePaginated(query, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search categories", zap.Error(err))
		return nil, 0, err
	}

	var responses []models.CategoryResponse
	for _, category := range categories {
		// Fetch subcategories for this category
		subcategories, err := s.subcategoryRepo.GetByCategoryID(category.ID)
		if err != nil {
			s.logger.Error("Failed to retrieve subcategories for category",
				zap.Error(err),
				zap.String("category_id", category.ID))
			return nil, 0, err
		}

		var subcatResponses []models.SubcategoryResponse
		for _, subcat := range subcategories {
			subcatResponse := models.SubcategoryResponse{
				ID:          subcat.ID,
				Name:        subcat.Name,
				Description: subcat.Description,
				CategoryID:  subcat.CategoryID,
				CreatedAt:   subcat.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:   subcat.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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

	s.logger.Info("Search categories completed",
		zap.Int("count", len(responses)),
		zap.Int64("total", total))
	return responses, total, nil
}

// GetAllCategoriesWithSubcategories retrieves all categories with their subcategories
// Note: Subcategories are fetched separately using GetByCategoryID since we removed
// the association from the Category model to reduce database load.
func (s *CategoryService) GetAllCategoriesWithSubcategories(ctx context.Context) ([]models.CategoryResponse, error) {
	s.logger.Info("Retrieving all categories with subcategories")

	categories, err := s.categoryRepo.GetAll()
	if err != nil {
		s.logger.Error("Failed to retrieve categories", zap.Error(err))
		return nil, err
	}

	var responses []models.CategoryResponse
	for _, category := range categories {
		// Fetch subcategories for this category using category ID
		subcategories, err := s.subcategoryRepo.GetByCategoryID(category.ID)
		if err != nil {
			s.logger.Error("Failed to retrieve subcategories for category",
				zap.Error(err),
				zap.String("category_id", category.ID))
			return nil, err
		}

		var subcatResponses []models.SubcategoryResponse
		for _, subcat := range subcategories {
			subcatResponse := models.SubcategoryResponse{
				ID:          subcat.ID,
				Name:        subcat.Name,
				Description: subcat.Description,
				CategoryID:  subcat.CategoryID,
				CreatedAt:   subcat.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:   subcat.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
