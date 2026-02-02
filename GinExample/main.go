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
	categoryService := services.NewCategoryService(models.DB)
	authService := services.NewAuthService(models.DB)
	limiter := utils.NewIPRateLimiter()

	// Initialize handlers
	articleHandler := handlers.NewArticleHandler(articleService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	authHandler := handlers.NewAuthHandler(authService)

	// Initialize Gin without default middleware
	r := gin.New()

	// Global Middleware
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logging())
	r.Use(middleware.RateLimit(limiter))
	r.Use(middleware.CORS("GET, POST, PUT, DELETE, OPTIONS"))

	// Public Routes
	r.GET("/health", authHandler.Health)
	r.POST("/auth/register", middleware.ContentType(), authHandler.Register)
	r.POST("/auth/login", middleware.ContentType(), authHandler.Login)
	r.GET("/ping", articleHandler.Ping)
	r.GET("/articles", articleHandler.GetArticles)
	r.GET("/articles/:id", articleHandler.GetArticle)
	r.GET("/categories", categoryHandler.GetCategories)
	r.GET("/categories/:id", categoryHandler.GetCategory)

	// Protected Routes (JWT required)
	protected := r.Group("/")
	protected.Use(middleware.JWTAuth())
	{
		protected.GET("/profile", authHandler.GetProfile)
		protected.PUT("/profile", middleware.ContentType(), authHandler.UpdateProfile)
		protected.PUT("/auth/change-password", middleware.ContentType(), authHandler.ChangePassword)
		protected.POST("/auth/refresh", authHandler.RefreshToken)
		protected.POST("/articles", middleware.ContentType(), articleHandler.CreateArticle)
		protected.PUT("/articles/:id", middleware.ContentType(), articleHandler.UpdateArticle)
		protected.DELETE("/articles/:id", articleHandler.DeleteArticle)
		protected.POST("/categories", middleware.ContentType(), categoryHandler.CreateCategory)
		protected.PUT("/categories/:id", middleware.ContentType(), categoryHandler.UpdateCategory)
		protected.DELETE("/categories/:id", categoryHandler.DeleteCategory)
	}

	// Admin Routes (JWT + Admin role required)
	admin := r.Group("/admin")
	admin.Use(middleware.JWTAuth(), middleware.AdminOnly())
	{
		admin.GET("/users", authHandler.GetAllUsers)
		admin.PUT("/users/:id/role", middleware.ContentType(), authHandler.UpdateUserRole)
		admin.GET("/stats", articleHandler.GetStats)
	}

	// Start server
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}