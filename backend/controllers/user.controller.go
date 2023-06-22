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
