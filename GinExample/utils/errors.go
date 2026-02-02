package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/models"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

// Error constructors
func NewBadRequestError(message, details string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: message, Details: details}
}

func NewNotFoundError(message string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: message}
}

func NewValidationError(details string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: "Validation failed", Details: details}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: message}
}

// Response helper
func HandleError(c *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		c.JSON(appErr.Code, models.APIResponse{
			Error:     appErr.Message,
			Message:   appErr.Details,
			RequestID: c.GetString("requestID"),
		})
	} else {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Error:     "Internal server error",
			RequestID: c.GetString("requestID"),
		})
	}
}

func Success(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		RequestID: c.GetString("requestID"),
	})
}

func Created(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		RequestID: c.GetString("requestID"),
	})
}