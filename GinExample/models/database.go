package models

import (
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:@tcp(localhost:3306)/STUDENT?charset=utf8mb4&parseTime=True&loc=Local"
	}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Drop existing tables to avoid constraint issues
	DB.Migrator().DropTable(&Article{}, &Category{})

	// Migrate categories first, then articles
	err = DB.AutoMigrate(&Category{}, &Article{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	seedData()
}

func seedData() {
	// Seed categories first
	var categoryCount int64
	DB.Model(&Category{}).Count(&categoryCount)
	
	if categoryCount == 0 {
		categories := []Category{
			{Name: "Technology", Description: "Articles about technology and programming"},
			{Name: "Web Development", Description: "Web development tutorials and guides"},
			{Name: "General", Description: "General articles and discussions"},
		}
		
		for _, category := range categories {
			DB.Create(&category)
		}
	}

	// Seed articles
	var articleCount int64
	DB.Model(&Article{}).Count(&articleCount)
	
	if articleCount == 0 {
		articles := []Article{
			{Title: "Getting Started with Go", Content: "Go is a programming language...", Author: "John Doe", CategoryID: 1},
			{Title: "Web Development with Gin", Content: "Gin is a web framework...", Author: "Jane Smith", CategoryID: 2},
		}
		
		for _, article := range articles {
			DB.Create(&article)
		}
	}
}