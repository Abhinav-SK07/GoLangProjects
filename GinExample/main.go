package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/handlers"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/middleware"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/services"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/utils"
	"github.com/Abhinav-SK07/GoLangProjects/GinExample/models"
)

func main() {
	// Initialize database
	models.InitDB()

	// Initialize services
	articleService := services.NewArticleService(models.DB)
	limiter := utils.NewIPRateLimiter()

	// Initialize handlers
	articleHandler := handlers.NewArticleHandler(articleService)

	// Initialize Gin without default middleware
	r := gin.New()

	// Global Middleware
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logging())
	r.Use(middleware.RateLimit(limiter))
	r.Use(middleware.CORS("GET, POST, PUT, DELETE, OPTIONS"))

	// Public Routes
	r.GET("/ping", articleHandler.Ping)
	r.GET("/articles", articleHandler.GetArticles)
	r.GET("/articles/:id", articleHandler.GetArticle)

	// Protected Routes Group
	protected := r.Group("/")
	protected.Use(middleware.Auth())
	{
		protected.POST("/articles", middleware.ContentType(), articleHandler.CreateArticle)
		protected.PUT("/articles/:id", middleware.ContentType(), articleHandler.UpdateArticle)
		protected.DELETE("/articles/:id", articleHandler.DeleteArticle)
		protected.GET("/admin/stats", articleHandler.GetStats)
	}

	// Start server
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}