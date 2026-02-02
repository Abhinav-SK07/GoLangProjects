package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/models"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/services"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/utils"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.AuthResponse{
			Success: false,
			Message: "Invalid input: " + err.Error(),
		})
		return
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	createdUser, err := h.service.Register(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.AuthResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	token, _ := utils.GenerateToken(createdUser.ID, createdUser.Username, createdUser.Role)

	c.JSON(http.StatusCreated, models.AuthResponse{
		Success: true,
		Token:   token,
		User:    *createdUser,
		Message: "User registered successfully",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var loginData struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, models.AuthResponse{
			Success: false,
			Message: "Username and password are required",
		})
		return
	}

	user, err := h.service.Login(loginData.Username, loginData.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.AuthResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	token, _ := utils.GenerateToken(user.ID, user.Username, user.Role)

	c.JSON(http.StatusOK, models.AuthResponse{
		Success: true,
		Token:   token,
		User:    *user,
		Message: "Login successful",
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt("userID")
	user, err := h.service.GetByID(userID)
	if err != nil {
		utils.HandleError(c, utils.NewNotFoundError("User not found"))
		return
	}
	utils.Success(c, user, "")
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetInt("userID")
	var updates models.User
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.HandleError(c, utils.NewBadRequestError("Invalid input", err.Error()))
		return
	}

	user, err := h.service.UpdateProfile(userID, updates)
	if err != nil {
		utils.HandleError(c, utils.NewBadRequestError("Update failed", err.Error()))
		return
	}

	utils.Success(c, user, "Profile updated successfully")
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID := c.GetInt("userID")
	username := c.GetString("username")
	role := c.GetString("role")

	token, err := utils.GenerateToken(userID, username, role)
	if err != nil {
		utils.HandleError(c, utils.NewBadRequestError("Token generation failed", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Success: true,
		Token:   token,
		Message: "Token refreshed successfully",
	})
}

func (h *AuthHandler) GetAllUsers(c *gin.Context) {
	users := h.service.GetAllUsers()
	utils.Success(c, users, "")
}

func (h *AuthHandler) UpdateUserRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.HandleError(c, utils.NewBadRequestError("Invalid user ID", ""))
		return
	}

	var roleData struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&roleData); err != nil {
		utils.HandleError(c, utils.NewBadRequestError("Role is required", ""))
		return
	}

	user, err := h.service.UpdateUserRole(id, roleData.Role)
	if err != nil {
		utils.HandleError(c, utils.NewBadRequestError("Update failed", err.Error()))
		return
	}

	utils.Success(c, user, "User role updated successfully")
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetInt("userID")
	var passwordData struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&passwordData); err != nil {
		utils.HandleError(c, utils.NewBadRequestError("Invalid input", err.Error()))
		return
	}

	if err := h.service.ChangePassword(userID, passwordData.OldPassword, passwordData.NewPassword); err != nil {
		utils.HandleError(c, utils.NewBadRequestError("Password change failed", err.Error()))
		return
	}

	utils.Success(c, nil, "Password changed successfully")
}

func (h *AuthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   "2024-01-01T00:00:00Z",
	})
}