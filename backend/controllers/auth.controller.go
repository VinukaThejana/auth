package controllers

import (
	"context"
	"time"

	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/errors"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/models"
	"github.com/VinukaThejana/auth/backend/schemas"
	"github.com/VinukaThejana/auth/backend/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Auth related controllers
type Auth struct{}

// Register is a function that is used to register a user with the backend
func (Auth) Register(c *fiber.Ctx, h *initialize.H) error {
	var payload *schemas.RegisterInput
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	if err := payload.Validate(); err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	if payload.Password != payload.PasswordConfirmation {
		return c.Status(fiber.StatusForbidden).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	newUser := models.User{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: string(hashedPassword),
	}

	result := h.DB.DB.Create(&newUser)
	if result.Error != nil {
		if result.Error == gorm.ErrDuplicatedKey {
			return c.Status(fiber.StatusBadRequest).JSON(response{
				Status: errors.ErrBadRequest.Error(),
			})
		}

		log.Error(result.Error, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response{
		Status: errors.Okay,
	})
}

