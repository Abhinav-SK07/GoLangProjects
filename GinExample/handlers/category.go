package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/models"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/services"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/utils"
)

type CategoryHandler struct {
	service *services.CategoryService
}

func NewCategoryHandler(service *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) GetCategories(c *gin.Context) {
	categories := h.service.GetAll()
	utils.Success(c, categories, "")
}

func (h *CategoryHandler) GetCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.HandleError(c, utils.NewBadRequestError("Invalid ID format", "ID must be a positive integer"))
		return
	}

	category, err := h.service.GetByID(id)
	if err != nil {
		utils.HandleError(c, utils.NewNotFoundError("Category not found"))
		return
	}

	utils.Success(c, category, "")
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var newCategory models.Category
	if err := c.ShouldBindJSON(&newCategory); err != nil {
		utils.HandleError(c, utils.NewBadRequestError("Invalid JSON format", err.Error()))
		return
	}

	category, err := h.service.Create(newCategory)
	if err != nil {
		utils.HandleError(c, utils.NewValidationError(err.Error()))
		return
	}

	utils.Created(c, category, "Category created successfully")
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.HandleError(c, utils.NewBadRequestError("Invalid ID format", "ID must be a positive integer"))
		return
	}

	var input models.Category
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.HandleError(c, utils.NewBadRequestError("Invalid JSON format", err.Error()))
		return
	}

	category, err := h.service.Update(id, input)
	if err != nil {
		if err.Error() == "category not found" {
			utils.HandleError(c, utils.NewNotFoundError("Category not found"))
		} else {
			utils.HandleError(c, utils.NewValidationError(err.Error()))
		}
		return
	}

	utils.Success(c, category, "Category updated successfully")
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.HandleError(c, utils.NewBadRequestError("Invalid ID format", "ID must be a positive integer"))
		return
	}

	if err := h.service.Delete(id); err != nil {
		if err.Error() == "category not found" {
			utils.HandleError(c, utils.NewNotFoundError("Category not found"))
		} else {
			utils.HandleError(c, utils.NewBadRequestError("Cannot delete category", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Category deleted successfully",
		RequestID: c.GetString("requestID"),
	})
}