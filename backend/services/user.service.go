package services

import (
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User struct contains user realated db operations
type User struct{}

// IsEmailAvailable is a function to check wether the email address given is already occupied
func (User) IsEmailAvailable(h *initialize.H, email string) (id *uuid.UUID, available bool, verified bool, err error) {
	var user models.User
	err = h.DB.DB.Select("id", "email", "verified").Where("email = ?", email).First(&user).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, false, false, err
		}

		return nil, true, false, err
	}

	return user.ID, false, *user.Verified, err
}

// IsUsernameAvailable is a function to check wether the given username is available in the database
func (User) IsUsernameAvailable(h *initialize.H, username string) (bool, error) {
	var user models.User
	err := h.DB.DB.Select("username").Where("username = ?", username).First(&user).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

// Create is a function that is used to create a user in the database
func (User) Create(h *initialize.H, user models.User) (newUser models.User, err error) {
	newUser = user
	err = h.DB.DB.Create(&newUser).Error
	if err != nil {
		return models.User{}, err
	}

	return newUser, nil
}
