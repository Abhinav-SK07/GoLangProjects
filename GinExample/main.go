package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Abhinav-SK07/GoLangProjects/GinExample/"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// -------------------------------------------------------------------
// Domain Models
// -------------------------------------------------------------------

// Article represents a blog article resource
type Article struct {
	ID        int       `json:"id"`
	Title     string    `json:"title" binding:"required"`
	Content   string    `json:"content" binding:"required"`
	Author    string    `json:"author" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIResponse represents a standardized JSON response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// -------------------------------------------------------------------
// Rate Limiter Structures
// -------------------------------------------------------------------

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	clients map[string]*client
	mu      sync.Mutex
}

// -------------------------------------------------------------------
// Global State
// -------------------------------------------------------------------

var (
	// Rate Limiter Instance
	limiterInstance = &IPRateLimiter{
		clients: make(map[string]*client),
	}

	// In-memory Database
	articles = []Article{
		{ID: 1, Title: "Getting Started with Go", Content: "Go is a programming language...", Author: "John Doe", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, Title: "Web Development with Gin", Content: "Gin is a web framework...", Author: "Jane Smith", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	// Database Mutex (Required for thread safety)
	mu sync.Mutex

	// ID Counter (Prevents collisions on delete/create cycles)
	nextID = 3
)

// -------------------------------------------------------------------
// Main Entry Point
// -------------------------------------------------------------------

func main() {
	// Initialize Gin without default middleware (we add our own)
	r := gin.New()

	// 1. Global Middleware (Applied to all requests)
	r.Use(ErrorHandlerMiddleware())
	r.Use(RequestIDMiddleware())
	r.Use(LoggingMiddleware())
	r.Use(RateLimitMiddleware())
	r.Use(CORSMiddleware("GET, POST, PUT, DELETE, OPTIONS"))

	// 2. Public Routes
	r.GET("/ping", ping)
	r.GET("/articles", getArticles)
	r.GET("/articles/:id", getArticle)

	// 3. Protected Routes Group
	protected := r.Group("/")
	protected.Use(AuthMiddleware()) // Requires X-API-Key header
	{
		protected.POST("/articles", ContentTypeMiddleware(), createArticle)
		protected.PUT("/articles/:id", ContentTypeMiddleware(), updateArticle)
		protected.DELETE("/articles/:id", deleteArticle)
		protected.GET("/admin/stats", getStats)
	}

	// Start server
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// -------------------------------------------------------------------
// Route Handlers
// -------------------------------------------------------------------

// GET /ping
func ping(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Message:   "pong",
		RequestID: c.GetString("requestID"),
	})
}

// GET /articles?page=1&limit=10
func getArticles(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	// Normalize parameters
	if page < 1 { page = 1 }
	if limit < 1 { limit = 10 }
	if limit > 100 { limit = 100 }

	mu.Lock()
	defer mu.Unlock()

	total := len(articles)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		// Page out of bounds, return empty list
		c.JSON(http.StatusOK, []Article{})
		return
	}
	if end > total {
		end = total
	}

	// Return a slice copy
	c.JSON(http.StatusOK, articles[start:end])
}

// GET /articles/:id
func getArticle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Error: "Invalid ID format"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	idx := findIndexByID(id)
	if idx == -1 {
		c.JSON(http.StatusNotFound, APIResponse{Error: "Article not found", Message: "404"})
		return
	}

	c.JSON(http.StatusOK, APIResponse{Success: true, Data: articles[idx]})
}

// POST /articles
func createArticle(c *gin.Context) {
	var newArticle Article
	if err := c.ShouldBindJSON(&newArticle); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Error: err.Error()})
		return
	}

	if err := validateArticle(newArticle); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Error: "Validation failed", Message: err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	newArticle.ID = nextID
	nextID++
	newArticle.CreatedAt = time.Now()
	newArticle.UpdatedAt = time.Now()

	articles = append(articles, newArticle)

	c.JSON(http.StatusCreated, APIResponse{Success: true, Data: newArticle})
}

// PUT /articles/:id
func updateArticle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Error: "Invalid ID format"})
		return
	}

	var input Article
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Error: err.Error()})
		return
	}

	if err := validateArticle(input); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Error: "Validation failed", Message: err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	idx := findIndexByID(id)
	if idx == -1 {
		c.JSON(http.StatusNotFound, APIResponse{Error: "Article not found", Message: "404"})
		return
	}

	// Update fields safely
	articles[idx].Title = input.Title
	articles[idx].Content = input.Content
	// Assuming Author might be updated, though typically this is restricted
	articles[idx].Author = input.Author 
	articles[idx].UpdatedAt = time.Now()

	c.JSON(http.StatusOK, APIResponse{Success: true, Data: articles[idx]})
}

// DELETE /articles/:id
func deleteArticle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Error: "Invalid ID format"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	idx := findIndexByID(id)
	if idx == -1 {
		c.JSON(http.StatusNotFound, APIResponse{Error: "Article not found", Message: "404"})
		return
	}

	// Remove from slice (performant delete preserving order)
	articles = append(articles[:idx], articles[idx+1:]...)

	c.JSON(http.StatusOK, APIResponse{Success: true, Message: "Article deleted successfully"})
}

// GET /admin/stats
func getStats(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.AbortWithStatusJSON(http.StatusForbidden, APIResponse{Error: "Forbidden", Message: "Admin access required"})
		return
	}

	mu.Lock()
	count := len(articles)
	mu.Unlock()

	stats := gin.H{
		"total_articles": count,
		"active_users":   5,   // Mock data
		"uptime_seconds": 360, // Mock data
	}

	c.JSON(http.StatusOK, APIResponse{Success: true, Data: stats})
}

// -------------------------------------------------------------------
// Middleware
// -------------------------------------------------------------------

func RequestIDMiddleware() gin.HandlerFunc {
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

func LoggingMiddleware() gin.HandlerFunc {
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

func AuthMiddleware() gin.HandlerFunc {
	validKeys := map[string]string{
		"admin-key-123": "admin",
		"user-key-456":  "user",
	}

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		role, exists := validKeys[apiKey]

		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, APIResponse{
				Error:   "Unauthorized",
				Message: "Invalid or missing API Key",
			})
			return
		}

		c.Set("role", role)
		c.Next()
	}
}

func ContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ct := c.GetHeader("Content-Type")
		// Use Contains to allow "application/json; charset=utf-8"
		if !strings.Contains(strings.ToLower(ct), "application/json") {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, APIResponse{
				Error:   "Unsupported Media Type",
				Message: "Content-Type must be application/json",
			})
			return
		}
		c.Next()
	}
}

func CORSMiddleware(allowedMethods string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Request-ID, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", allowedMethods)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, APIResponse{
					Error:   "Internal Server Error",
					Message: "Something went wrong",
				})
			}
		}()
		c.Next()
	}
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		l := getIPLimiter(ip)

		if !l.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, APIResponse{
				Error: "Rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

// -------------------------------------------------------------------
// Helper Functions
// -------------------------------------------------------------------

// getIPLimiter manages rate limiters for individual IPs
func getIPLimiter(ip string) *rate.Limiter {
	limiterInstance.mu.Lock()
	defer limiterInstance.mu.Unlock()

	v, exists := limiterInstance.clients[ip]
	if !exists {
		// Rate limit: 100 requests per minute
		limiter := rate.NewLimiter(rate.Every(time.Minute/100), 100)
		limiterInstance.clients[ip] = &client{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// findIndexByID locates an article's index in the slice (Must be called within a Mutex lock)
func findIndexByID(id int) int {
	for i, a := range articles {
		if a.ID == id {
			return i
		}
	}
	return -1
}

// validateArticle checks input requirements and security keywords
func validateArticle(a Article) error {
	if len(strings.TrimSpace(a.Title)) == 0 || len(strings.TrimSpace(a.Content)) == 0 || len(strings.TrimSpace(a.Author)) == 0 {
		return fmt.Errorf("title, content, and author are required")
	}
	if len(a.Title) > 200 {
		return fmt.Errorf("title too long (max 200 chars)")
	}
	if len(a.Content) < 10 {
		return fmt.Errorf("content too short (min 10 chars)")
	}

	// Basic security keyword check (Naive implementation for demo purposes)
	lowerContent := strings.ToLower(a.Content)
	forbidden := []string{"<script>", "javascript:", "drop table", "select * from", "malware", "exploit"}
	
	for _, word := range forbidden {
		if strings.Contains(lowerContent, word) {
			return fmt.Errorf("content contains forbidden keyword: %s", word)
		}
	}
	
	return nil
}