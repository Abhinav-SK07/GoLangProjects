package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/models"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/utils"
	"github.com/google/uuid"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("requestID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		duration := time.Since(t)
		reqID := c.GetString("requestID")
		status := c.Writer.Status()

		log.Printf("[GIN] ID=%s | Status=%d | Method=%s | Path=%s | Duration=%v",
			reqID, status, c.Request.Method, c.Request.URL.Path, duration)
	}
}

func Auth() gin.HandlerFunc {
	validKeys := map[string]string{
		"admin-key-123": "admin",
		"user-key-456":  "user",
	}

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Error:     "Unauthorized",
				Message:   "API Key is required",
				RequestID: c.GetString("requestID"),
			})
			return
		}

		role, exists := validKeys[apiKey]
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Error:     "Unauthorized",
				Message:   "Invalid API Key",
				RequestID: c.GetString("requestID"),
			})
			return
		}

		c.Set("role", role)
		c.Next()
	}
}

func ContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		ct := c.GetHeader("Content-Type")
		if !strings.Contains(strings.ToLower(ct), "application/json") {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, models.APIResponse{
				Error:   "Unsupported Media Type",
				Message: "Content-Type must be application/json",
			})
			return
		}
		c.Next()
	}
}

func RateLimit(limiter *utils.IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.GetLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, models.APIResponse{
				Error:   "Rate limit exceeded",
				Message: "Too many requests",
			})
			return
		}
		c.Next()
	}
}

func CORS(allowedMethods string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Request-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, models.APIResponse{
					Error:     "Internal server error",
					Message:   "An unexpected error occurred",
					RequestID: c.GetString("requestID"),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}