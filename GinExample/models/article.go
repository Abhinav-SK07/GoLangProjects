package models

import "time"

// Article represents a blog article resource
type Article struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement" db:"id"`

	Title     string    `json:"title" binding:"required,max=100" gorm:"not null;size:100"`
	Content   string    `json:"content" binding:"required,max=250" gorm:"not null;size:250"`
	Author    string    `json:"author" binding:"required,max=100" gorm:"not null;size:100"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// APIResponse represents a standardized JSON response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}