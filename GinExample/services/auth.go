package services

import (
	"errors"
	"strings"

	"github.com/Abhinav-SK07/GoLangProjects/GinExample/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Register(user models.User) (*models.User, error) {
	if err := ValidateUser(user); err != nil {
		return nil, err
	}

	// Check if username or email exists
	var existingUser models.User
	if err := s.db.Where("username = ? OR email = ?", user.Username, user.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("username or email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(hashedPassword)
	user.Role = "user"
	user.Active = true

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) Login(username, password string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("username = ? AND active = ?", username, true).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

func (s *AuthService) GetByID(id int) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (s *AuthService) UpdateProfile(id int, updates models.User) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if updates.Email != "" {
		user.Email = updates.Email
	}
	if updates.Username != "" {
		user.Username = updates.Username
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) GetAllUsers() []models.User {
	var users []models.User
	s.db.Find(&users)
	return users
}

func (s *AuthService) UpdateUserRole(id int, role string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if role != "user" && role != "admin" {
		return nil, errors.New("invalid role")
	}

	user.Role = role
	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) ChangePassword(userID int, oldPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	if len(newPassword) < 6 {
		return errors.New("new password must be at least 6 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return s.db.Save(&user).Error
}

func ValidateUser(user models.User) error {
	if strings.TrimSpace(user.Username) == "" {
		return errors.New("username is required")
	}
	if len(user.Username) < 3 || len(user.Username) > 20 {
		return errors.New("username must be between 3 and 20 characters")
	}
	if strings.TrimSpace(user.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(user.Password) == "" {
		return errors.New("password is required")
	}
	if len(user.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}