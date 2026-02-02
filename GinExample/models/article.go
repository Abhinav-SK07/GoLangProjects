package model

import (
    "time"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

type Article struct {
    ID        int       `json:"id" gorm:"primaryKey"`
    Title     string    `json:"title" binding:"required"`
    Content   string    `json:"content" binding:"required"`
    Author    string    `json:"author" binding:"required"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

var DB *gorm.DB

func ConnectDatabase() {
    database, err := gorm.Open(mysql.Open("articles.db"), &gorm.Config{})
    if err != nil {
        panic("Failed to connect to database!")
    }

    database.AutoMigrate(&Article{})
    DB = database
}