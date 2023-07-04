package services

import (
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/models"
	"gorm.io/gorm"
)

// User struct contains user realated db operations
type User struct{}

// IsEmailAvailable is a function to check wether the email address given is already occupied
func (User) IsEmailAvailable(h *initialize.H, email string) (bool, error) {
	var user models.User
	err := h.DB.DB.Select("email").Where("email = ?", email).First(&user).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return false, err
		}

		return true, nil
	}

	return false, nil
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
