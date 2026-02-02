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

	err = DB.AutoMigrate(&Article{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	seedData()
}

func seedData() {
	var count int64
	DB.Model(&Article{}).Count(&count)
	
	if count == 0 {
		articles := []Article{
			{Title: "Getting Started with Go", Content: "Go is a programming language...", Author: "John Doe"},
			{Title: "Web Development with Gin", Content: "Gin is a web framework...", Author: "Jane Smith"},
		}
		
		for _, article := range articles {
			DB.Create(&article)
		}
	}
}