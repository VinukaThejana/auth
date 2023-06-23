package controllers

import (
	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/errors"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/models"
	"github.com/VinukaThejana/auth/backend/schemas"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// User is struct contaning the user related controllers
type User struct{}

// GetUser is a function that is used to get the details of the user
func (User) GetUser(c *fiber.Ctx, h *initialize.H) error {
	user := c.Locals(config.Enums{}.USER()).(schemas.User)
	var payload models.User
	if err := h.DB.DB.First(&payload, "id = ?", user.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(response{
				Status: errors.ErrUnauthorized.Error(),
			})
		}

		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"user": schemas.FilterUserRecord(&payload),
	})
}

// ValidateUsername is a function that is used to validate the username of the user
func (User) ValidateUsername(c *fiber.Ctx, h *initialize.H) error {
	var payload struct {
		Username string `json:"username"`
	}
	if err := c.BodyParser(&payload); err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	type response struct {
		Available bool `json:"available"`
	}

	var user models.User
	result := h.DB.DB.First(&user, "username = ?", payload.Username)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusOK).JSON(response{
				Available: true,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Available: false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response{
		Available: false,
	})
}
