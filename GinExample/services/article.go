package services

import (
	"errors"
	"strings"

	"github.com/Abhinav-SK07/GoLangProjects/GinExample/models"
	"gorm.io/gorm"
)

type ArticleService struct {
	db *gorm.DB
}

func NewArticleService(db *gorm.DB) *ArticleService {
	return &ArticleService{db: db}
}

func (s *ArticleService) GetAll(page, limit int) []models.Article {
	var articles []models.Article
	offset := (page - 1) * limit
	s.db.Preload("Category").Offset(offset).Limit(limit).Find(&articles)
	return articles
}

func (s *ArticleService) GetByID(id int) (*models.Article, error) {
	var article models.Article
	result := s.db.Preload("Category").First(&article, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("article not found")
		}
		return nil, result.Error
	}
	return &article, nil
}

func (s *ArticleService) Create(article models.Article) (*models.Article, error) {
	// Validate category exists
	var category models.Category
	result := s.db.First(&category, article.CategoryID)
	if result.Error != nil {
		return nil, errors.New("invalid category ID")
	}
	
	s.db.Create(&article)
	s.db.Preload("Category").First(&article, article.ID)
	return &article, nil
}

func (s *ArticleService) Update(id int, input models.Article) (*models.Article, error) {
	var article models.Article
	result := s.db.First(&article, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("article not found")
		}
		return nil, result.Error
	}

	// Validate category exists if provided
	if input.CategoryID != 0 {
		var category models.Category
		result := s.db.First(&category, input.CategoryID)
		if result.Error != nil {
			return nil, errors.New("invalid category ID")
		}
		article.CategoryID = input.CategoryID
	}

	article.Title = input.Title
	article.Content = input.Content
	article.Author = input.Author
	s.db.Save(&article)
	s.db.Preload("Category").First(&article, article.ID)
	return &article, nil
}

func (s *ArticleService) Delete(id int) error {
	result := s.db.Delete(&models.Article{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("article not found")
	}
	return nil
}

func (s *ArticleService) Count() int {
	var count int64
	s.db.Model(&models.Article{}).Count(&count)
	return int(count)
}

func ValidateArticle(article models.Article) error {
	if strings.TrimSpace(article.Title) == "" {
		return errors.New("title is required")
	}
	if len(article.Title) > 100 {
		return errors.New("title must be less than 100 characters")
	}
	if strings.TrimSpace(article.Content) == "" {
		return errors.New("content is required")
	}
	if len(article.Content) > 250 {
		return errors.New("content must be less than 250 characters")
	}
	if strings.TrimSpace(article.Author) == "" {
		return errors.New("author is required")
	}
	if len(article.Author) > 100 {
		return errors.New("author must be less than 100 characters")
	}
	if article.CategoryID <= 0 {
		return errors.New("category ID is required")
	}
	return nil
}