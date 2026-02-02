package services

import (
	"errors"
	"strings"

	"github.com/Abhinav-SK07/GoLangProjects/GinExample/models"
	"gorm.io/gorm"
)

type CategoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{db: db}
}

func (s *CategoryService) GetAll() []models.Category {
	var categories []models.Category
	s.db.Find(&categories)
	return categories
}

func (s *CategoryService) GetByID(id int) (*models.Category, error) {
	var category models.Category
	result := s.db.First(&category, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, result.Error
	}
	return &category, nil
}

func (s *CategoryService) Create(category models.Category) (*models.Category, error) {
	if err := ValidateCategory(category); err != nil {
		return nil, err
	}
	
	result := s.db.Create(&category)
	if result.Error != nil {
		return nil, result.Error
	}
	return &category, nil
}

func (s *CategoryService) Update(id int, input models.Category) (*models.Category, error) {
	var category models.Category
	result := s.db.First(&category, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, result.Error
	}

	if err := ValidateCategory(input); err != nil {
		return nil, err
	}

	category.Name = input.Name
	category.Description = input.Description
	s.db.Save(&category)
	return &category, nil
}

func (s *CategoryService) Delete(id int) error {
	// Check if category has articles
	var count int64
	s.db.Model(&models.Article{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("cannot delete category with existing articles")
	}

	result := s.db.Delete(&models.Category{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("category not found")
	}
	return nil
}

func ValidateCategory(category models.Category) error {
	if strings.TrimSpace(category.Name) == "" {
		return errors.New("name is required")
	}
	if len(category.Name) > 50 {
		return errors.New("name must be less than 50 characters")
	}
	if len(category.Description) > 200 {
		return errors.New("description must be less than 200 characters")
	}
	return nil
}