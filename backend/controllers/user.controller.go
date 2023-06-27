package controllers

import (
	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/errors"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/models"
	"github.com/VinukaThejana/auth/backend/schemas"
	"github.com/VinukaThejana/auth/backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User is struct contaning the user related controllers
type User struct{}

// GetUser is a function that is used to get the details of the user
func (User) GetUser(c *fiber.Ctx, h *initialize.H) error {
	userID := c.Locals(config.Enums{}.USER()).(string)
	var payload models.User
	if err := h.DB.DB.First(&payload, "id = ?", userID).Error; err != nil {
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

// UpdateEmail is a function that is used to update the users email address
func (User) UpdateEmail(c *fiber.Ctx, h *initialize.H, env *config.Env) error {
	var payload struct {
		Email string `json:"email" validate:"required,email"`
	}
	if err := c.BodyParser(&payload); err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	var user models.User
	result := h.DB.DB.First(&user, "email = ?", payload.Email)
	if result.Error == nil {
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrEmailAlreadyUsed.Error(),
		})
	}
	if result.Error != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	if ok := log.Validate(payload); !ok {
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	userID, err := uuid.Parse(c.Locals(config.Enums{}.USER()).(string))
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	verified := false
	err = h.DB.DB.Model(&models.User{}).Where("id = ?", userID.String()).Updates(models.User{
		Email:    payload.Email,
		Verified: &verified,
	}).Error
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusOK).JSON(response{
			Status: errors.Okay,
		})
	}

	go utils.Email{}.SendConfirmation(h, env, payload.Email, userID.String())

	return c.Status(fiber.StatusOK).JSON(response{
		Status: errors.Okay,
	})
}

// UpdateUsername is a function that is used to update the username of the user
func (User) UpdateUsername(c *fiber.Ctx, h *initialize.H) error {
	var payload struct {
		Username string `json:"username" validate:"required,min=3,max=15"`
	}

	if err := c.BodyParser(&payload); err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}
	if ok := log.Validate(payload); !ok {
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	var user models.User
	result := h.DB.DB.First(&user, "username = ?", payload.Username)
	if result.Error == nil {
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrUsernameAlreadyUsed.Error(),
		})
	}
	if result.Error != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	userID, err := uuid.Parse(c.Locals(config.Enums{}.USER()).(string))
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	err = h.DB.DB.Model(&models.User{}).Where("id = ?", userID).Update("username", payload.Username).Error
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusOK).JSON(response{
			Status: errors.Okay,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response{
		Status: errors.Okay,
	})
}

// UpdateName is a function that is used to update the name of the user
func (User) UpdateName(c *fiber.Ctx, h *initialize.H) error {
	var payload struct {
		Name string `json:"name" validate:"required,min=3,max=30"`
	}

	if err := c.BodyParser(&payload); err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}
	if ok := log.Validate(payload); !ok {
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	userID, err := uuid.Parse(c.Locals(config.Enums{}.USER()).(string))
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	err = h.DB.DB.Model(&models.User{}).Where("id = ?", userID).Update("name", payload.Name).Error
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusOK).JSON(response{
			Status: errors.Okay,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response{
		Status: errors.Okay,
	})
}
