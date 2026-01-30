package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// Article represents a blog article
type Article struct {
	ID        int       `json:"id"`
	Title     string    `json:"title" binding:"required"`
	Content   string    `json:"content" binding:"required"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	clients map[string]*client
	mu      sync.Mutex
}

var limiterInstance = &IPRateLimiter{
	clients: make(map[string]*client),
}

// In-memory storage
var (
	articles = []Article{
	{ID: 1, Title: "Getting Started with Go", Content: "Go is a programming language...", Author: "John Doe", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: 2, Title: "Web Development with Gin", Content: "Gin is a web framework...", Author: "Jane Smith", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	mu sync.Mutex
)
var nextID = 3

func main() {
	// TODO: Create Gin router without default middleware
	// Use gin.New() instead of gin.Default()
	r := gin.New()

	// TOP: Global protection
	r.Use(ErrorHandlerMiddleware(), RequestIDMiddleware(), LoggingMiddleware())
	r.Use(RateLimitMiddleware(), ContentTypeMiddleware())

// MIDDLE: Defined Groups
	protected := r.Group("/api").Use(AuthMiddleware())

    // 2. Route Groups SECOND
    protected.GET("/resource", resource)

	// Public routes (no authentication required)
	r.GET("/ping", ping)
	r.GET("/articles", getArticles)
	r.GET("/articles/:id", getArticle)
	// Protected routes (require authentication)
	r.POST("/articles", AuthMiddleware(), createArticle)
	r.PUT("/articles/:id", AuthMiddleware(), updateArticle)
	r.DELETE("/articles/:id", AuthMiddleware(), deleteArticle)
	r.GET("/admin/stats", AuthMiddleware(), getStats)

	// TODO: Define routes
	r.Run(":8080")
	// Public: GET /ping, GET /articles, GET /articles/:id
	
	// Protected: POST /articles, PUT /articles/:id, DELETE /articles/:id, GET /admin/stats

	// TODO: Start server on port 8080
}

// TODO: Implement middleware functions

// RequestIDMiddleware generates a unique request ID for each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set the ID in the Gin context for access in other handlers
		c.Set("requestID", requestID)

		// Set the ID in the response header
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next() // Pass control to the next handler
	}
}

// LoggingMiddleware logs all requests with timing information
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		userAgent := c.Request.Header.Get("User-Agent")
		r_id := c.MustGet("requestID").(string)


		// --- Before request ---
		fmt.Printf("Request started: Request ID %s Method %s Path %s Started At %s IP %s Agent %s\n",
					r_id, c.Request.Method, c.Request.URL.Path, t, c.ClientIP(), userAgent)

		// Pass to next handler in the chain
		c.Next()

		// --- After request ---
		duration := time.Since(t)
		status := c.Writer.Status()
		fmt.Printf("Request completed: Request ID %s Method %s Path %s Status %d Duration %s IP %s Agent %s\n",
					r_id, c.Request.Method, c.Request.URL.Path, status, duration, c.ClientIP(), userAgent)
	}
}

// AuthMiddleware validates API keys for protected routes
func AuthMiddleware() gin.HandlerFunc {
	// TODO: Define valid API keys and their roles
	// "admin-key-123" -> "admin"
	// "user-key-456" -> "user"
	keyToRole := map[string]string{
		"admin-key-123": "admin",
		"user-key-456":  "user",
	}

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		// Validate key existence in our map
		role, exists := keyToRole[apiKey]
		if apiKey == "" || !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, APIResponse{
				Error:   "401",
				Message: "Unauthorized: Invalid or missing API key",
			})
			return
		}

		// Store both key and role for downstream middlewares (like AdminOnly)
		c.Set("api-key", apiKey)
		c.Set("role", role)

		c.Next()
	}
}

// Example handler for a protected route
func resource(c *gin.Context) {
	// 1. Retrieve data from the Gin context
	// Using MustGet is safe if the middleware guarantees these keys exist
	role := c.GetString("role")

	// 2. Respond with user-specific context
	c.JSON(http.StatusOK, APIResponse{
		Success: true, 
		Message: role})
}

func CORSMiddleware(allowedMethods string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Request-ID, Authorization")
		
		// Set the methods based on the input string
		c.Writer.Header().Set("Access-Control-Allow-Methods", allowedMethods)

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func GetLimiter(ip string) *rate.Limiter {
	limiterInstance.mu.Lock()
	defer limiterInstance.mu.Unlock()

	c, exists := limiterInstance.clients[ip]
	if !exists {
		// 100 requests per minute
		limit := rate.Every(time.Minute / 100)
		l := rate.NewLimiter(limit, 100)
		limiterInstance.clients[ip] = &client{l, time.Now()}
		return l
	}

	c.lastSeen = time.Now()
	return c.limiter
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		l := GetLimiter(ip)

		// Check the state without consuming a token first
		now := time.Now()
		rv := l.ReserveN(now, 1)

		if !rv.OK() {
			// This usually happens if the burst is exceeded
			c.AbortWithStatusJSON(http.StatusTooManyRequests, APIResponse{Error: "Rate limit exceeded"})
			return
		}

		// Calculate headers
		delay := rv.DelayFrom(now)
		limit := "100"
		remaining := int(math.Max(0, float64(l.Tokens())))
		reset := now.Add(delay).Unix()

		c.Writer.Header().Set("X-RateLimit-Limit", limit)
		c.Writer.Header().Set("X-RateLimit-Remaining", string(rune(remaining)))
		c.Writer.Header().Set("X-RateLimit-Reset", string(rune(reset)))

		if delay > 0 {
			// If we had to "wait" for this token, it means we've hit the limit
			rv.Cancel() // Do not consume the token since we are blocking the request
			c.AbortWithStatusJSON(http.StatusTooManyRequests, APIResponse{
				Error: "Rate limit exceeded. Try again later.",
			})
			return
		}

		c.Next()
	}
}

// ContentTypeMiddleware validates content type for POST/PUT requests
func ContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Check content type for POST/PUT requests
		// Must be application/json
		// Return 415 if invalid content type
		method := c.Request.Method
		
		// Only validate methods that carry a request body
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
			contentType := c.GetHeader("Content-Type")

			// Check if Content-Type is missing or not application/json
			if contentType == "" || !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
				c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, APIResponse{
					Message: "Unsupported Media Type. Please use application/json",
					Error:   "415",
				})
				return
			}
		}
		c.Next()
	}
}

// ErrorHandlerMiddleware handles panics and errors
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Recovery from panics
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic: %v", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, APIResponse{
					Success: false,
					Error:   "500",
					Message: "Internal Server Error",
				})
			}
		}()

		c.Next()

		// Handle logical errors attached via c.Error()
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			c.JSON(c.Writer.Status(), APIResponse{
				Success: false,
				Message: err.Error(),
				Error:   strconv.Itoa(c.Writer.Status()),
			})
		}
	}
}


// TODO: Implement route handlers

// ping handles GET /ping - health check endpoint
func ping(c *gin.Context) {
	// TODO: Return simple pong response with request ID
		c.JSON(http.StatusOK, APIResponse{
		Success:  true,
		Message: "pong",
		RequestID: c.GetString("requestID"),
	})
}

// getArticles handles GET /articles - get all articles with pagination
func getArticles(c *gin.Context) {
	// TODO: Implement pagination (optional)
	// TODO: Return articles in standard format
	c.JSON(http.StatusOK, articles)
}

// getArticle handles GET /articles/:id - get article by ID
func getArticle(c *gin.Context) {
	// TODO: Get article ID from URL parameter
	// TODO: Find article by ID
	// TODO: Return 404 if not found
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Error: "Invalid article ID"})
		return
	}
	article, _ := findArticleByID(id)
	if article == nil {
		c.JSON(http.StatusNotFound, APIResponse{Message: "Article not found", Error: "404"})
		return
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: article})
}

// createArticle handles POST /articles - create new article (protected)
func createArticle(c *gin.Context) {
	// TODO: Parse JSON request body
	// TODO: Validate required fields
	// TODO: Add article to storage
	// TODO: Return created article
	var newArticle Article

	// 1. Parse & Validate JSON request body
	// ShouldBindJSON uses 'binding' tags and returns 400 automatically if validation fails
	if err := c.ShouldBindJSON(&newArticle); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Error: "400"})
		return
	}

	// 2. Add article to storage (Thread-safe)
	mu.Lock()
	newArticle.ID = len(articles) + 1
	articles = append(articles, newArticle)
	mu.Unlock()

	// 3. Return created article with 201 Created
	c.JSON(http.StatusCreated, newArticle)
}

// updateArticle handles PUT /articles/:id - update article (protected)
func updateArticle(c *gin.Context) {
	// TODO: Get article ID from URL parameter
	// TODO: Parse JSON request body
	// TODO: Find and update article
	// TODO: Return updated article
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Error: "Invalid article ID"})
		return
	}
	
	var updatedArticle Article
	// 1. Parse & Validate JSON request body
	if err := c.ShouldBindJSON(&updatedArticle); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Error: "400"})
		return
	}
	
	// 2. Find and update article (Thread-safe)
	mu.Lock()
	article, index := findArticleByID(id)
	if article == nil {
		mu.Unlock()
		c.JSON(http.StatusNotFound, APIResponse{Message: "Article not found", Error: "404"})
		return
	}
	// Update fields
	article.Title = updatedArticle.Title
	article.Content = updatedArticle.Content
	article.Author = updatedArticle.Author
	article.UpdatedAt = time.Now()
	articles[index] = *article
	mu.Unlock()
	
	// 3. Return updated article
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: article})
}

// deleteArticle handles DELETE /articles/:id - delete article (protected)
func deleteArticle(c *gin.Context) {
	// TODO: Get article ID from URL parameter
	// TODO: Find and remove article
	// TODO: Return success message
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Error: "Invalid article ID"})
		return
	}

	// 2. Find and remove article (Thread-safe)
	mu.Lock()
	_, index := findArticleByID(id)
	if index == -1 {
		mu.Unlock()
		c.JSON(http.StatusNotFound, APIResponse{Message: "Article not found", Error: "404"})
		return
	}
	// Remove article from slice
	articles = append(articles[:index], articles[index+1:]...)
	mu.Unlock()
	
	// 3. Return success message
	c.JSON(http.StatusOK, APIResponse{Success: true, Message: "Article deleted successfully"})
}

// getStats handles GET /admin/stats - get API usage statistics (admin only)
func getStats(c *gin.Context) {
	// 1. Check if user role is "admin"
	role, _ := c.Get("role")
	if role != "admin" {
		c.AbortWithStatusJSON(http.StatusForbidden, APIResponse{
			Success: false,
			Message: "Forbidden: Admin access required",
			Error:   "403",
		})
		return
	}

	// 2. Prepare mock statistics
	// Note: 'articles' should be the global slice defined in your storage logic
	stats := gin.H{
		"total_articles": len(articles),
		"total_requests": 1250, // Mock value
		"uptime":         "24h",
		"active_users":   5,
	}

	// 3. Return stats in standard format
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Statistics retrieved successfully",
		Data:    stats,
	})
}

// Helper functions

// findArticleByID finds an article by ID
func findArticleByID(id int) (*Article, int) {
	// TODO: Implement article lookup
	// Return article pointer and index, or nil and -1 if not found
	for i, article := range articles {
		if article.ID == id {
			return &articles[i], i
		}
	}
	return nil, -1
}

// validateArticle validates article data
func validateArticle(article Article) error {
	// TODO: Implement validation
	// Check required fields: Title, Content, Author
	if article.Title == "" || article.Content == "" || article.Author == "" {
		return fmt.Errorf("Title, Content, and Author are required")
	}
	if len(article.Content) < 10 {
		return fmt.Errorf("Content must be at least 10 characters long")
	}
	if len(article.Title) > 200 {
		return fmt.Errorf("Title must be less than 200 characters long")
	}
	if len(article.Author) > 100 {
		return fmt.Errorf("Author name must be less than 100 characters long")
	}
	if article.Author == "admin" {
		return fmt.Errorf("Author name cannot be 'admin'")
	}
	if strings.Contains(strings.ToLower(article.Title), "spam") {
		return fmt.Errorf("Title cannot contain spam keywords")
	}
	if strings.Contains(strings.ToLower(article.Content), "spam") {
		return fmt.Errorf("Content cannot contain spam keywords")
	}
	if strings.Contains(strings.ToLower(article.Content), "http") && !strings.Contains(strings.ToLower(article.Content), "https") {
		return fmt.Errorf("Content cannot contain unsecure links")
	}
	if strings.Contains(strings.ToLower(article.Content), "malware") {
		return fmt.Errorf("Content cannot contain malware keywords")
	}
	if strings.Contains(strings.ToLower(article.Content), "hack") {
		return fmt.Errorf("Content cannot contain hacking keywords")
	}
	if strings.Contains(strings.ToLower(article.Content), "sql injection") {
		return fmt.Errorf("Content cannot contain SQL injection keywords")
	}
	if strings.Contains(strings.ToLower(article.Content), "xss") {
		return fmt.Errorf("Content cannot contain XSS keywords")
	}
	if strings.Contains(strings.ToLower(article.Content), "phishing") {
		return fmt.Errorf("Content cannot contain phishing keywords")
	}
	if strings.Contains(strings.ToLower(article.Content), "ddos") {
		return fmt.Errorf("Content cannot contain DDoS keywords")
	}
	if strings.Contains(strings.ToLower(article.Content), "botnet") {
		return fmt.Errorf("Content cannot contain botnet keywords")
	}
	if strings.Contains(strings.ToLower(article.Content), "ransomware") {
		return fmt.Errorf("Content cannot contain ransomware keywords")
	}
	if strings.Contains(strings.ToLower(article.Content), "exploit") {
		return fmt.Errorf("Content cannot contain exploit keywords")
	}
	return nil
}