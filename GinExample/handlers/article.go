package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/models"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/services"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/utils"
)

type ArticleHandler struct {
	service *services.ArticleService
}

func NewArticleHandler(service *services.ArticleService) *ArticleHandler {
	return &ArticleHandler{service: service}
}

func (h *ArticleHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "pong",
		RequestID: c.GetString("requestID"),
	})
}

func (h *ArticleHandler) GetArticles(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Error:     "Invalid page parameter",
			RequestID: c.GetString("requestID"),
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Error:     "Invalid limit parameter",
			RequestID: c.GetString("requestID"),
		})
		return
	}

	if limit > 100 {
		limit = 100
	}

	articles := h.service.GetAll(page, limit)
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      articles,
		RequestID: c.GetString("requestID"),
	})
}

func (h *ArticleHandler) GetArticle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.HandleError(c, utils.NewBadRequestError("Invalid ID format", "ID must be a positive integer"))
		return
	}

	article, err := h.service.GetByID(id)
	if err != nil {
		utils.HandleError(c, utils.NewNotFoundError("Article not found"))
		return
	}

	utils.Success(c, article, "")
}

func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	var newArticle models.Article
	if err := c.ShouldBindJSON(&newArticle); err != nil {
		utils.HandleError(c, utils.NewBadRequestError("Invalid JSON format", err.Error()))
		return
	}

	if err := services.ValidateArticle(newArticle); err != nil {
		utils.HandleError(c, utils.NewValidationError(err.Error()))
		return
	}

	article := h.service.Create(newArticle)
	utils.Created(c, article, "Article created successfully")
}

func (h *ArticleHandler) UpdateArticle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Error:     "Invalid ID format - must be positive integer",
			RequestID: c.GetString("requestID"),
		})
		return
	}

	var input models.Article
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Error:     "Invalid JSON format or validation failed",
			Message:   err.Error(),
			RequestID: c.GetString("requestID"),
		})
		return
	}

	if err := services.ValidateArticle(input); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Error:     "Validation failed",
			Message:   err.Error(),
			RequestID: c.GetString("requestID"),
		})
		return
	}

	article, err := h.service.Update(id, input)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Error:     "Article not found",
			RequestID: c.GetString("requestID"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      article,
		Message:   "Article updated successfully",
		RequestID: c.GetString("requestID"),
	})
}

func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Error:     "Invalid ID format - must be positive integer",
			RequestID: c.GetString("requestID"),
		})
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Error:     "Article not found",
			RequestID: c.GetString("requestID"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Article deleted successfully",
		RequestID: c.GetString("requestID"),
	})
}

func (h *ArticleHandler) GetStats(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.AbortWithStatusJSON(http.StatusForbidden, models.APIResponse{
			Error:     "Forbidden",
			Message:   "Admin access required",
			RequestID: c.GetString("requestID"),
		})
		return
	}

	count := h.service.Count()
	stats := gin.H{
		"total_articles": count,
		"active_users":   5,
		"uptime_seconds": 360,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      stats,
		RequestID: c.GetString("requestID"),
	})
}