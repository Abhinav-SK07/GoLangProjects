package models

import "time"

// User represents a user in the system
type User struct {
	ID       int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Username string    `json:"username" gorm:"not null;size:20;uniqueIndex"`
	Email    string    `json:"email" gorm:"not null;size:100;uniqueIndex"`
	Password string    `json:"-" gorm:"not null;size:255"`
	Role     string    `json:"role" gorm:"not null;default:'user';size:10"`
	Active   bool      `json:"active" gorm:"not null;default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// RegisterRequest for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	User    User   `json:"user,omitempty"`
	Message string `json:"message,omitempty"`
}

// Category represents an article category
type Category struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"not null;size:50;uniqueIndex"`
	Description string    `json:"description" gorm:"size:200"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Article represents a blog article resource
type Article struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement" db:"id"`
	Title      string    `json:"title" binding:"required,max=100" gorm:"not null;size:100"`
	Content    string    `json:"content" binding:"required,max=250" gorm:"not null;size:250"`
	Author     string    `json:"author" binding:"required,max=100" gorm:"not null;size:100"`
	CategoryID int       `json:"category_id" binding:"required" gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:CategoryID;references:ID"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// APIResponse represents a standardized JSON response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}