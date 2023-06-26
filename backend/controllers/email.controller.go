package controllers

import (
	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/errors"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/utils"
	"github.com/gofiber/fiber/v2"
)

// Email all email related controllers
type Email struct{}

// ConfirmEmail is a function that is used to confirm the email address
func (Email) ConfirmEmail(c *fiber.Ctx, h *initialize.H, env *config.Env) error {
	token := c.Query("token")
	userID := c.Locals(config.Enums{}.USER()).(string)

	err := utils.Email{}.ConfirmEmail(h, token, userID)
	if err != nil {
		switch err {
		case errors.ErrEmailConfirmationExpired:
			return c.Status(fiber.StatusBadRequest).JSON(response{
				Status: err.Error(),
			})
		case errors.ErrBadRequest:
			return c.Status(fiber.StatusBadRequest).JSON(response{
				Status: err.Error(),
			})
		case errors.ErrUnauthorized:
			return c.Status(fiber.StatusUnauthorized).JSON(response{
				Status: err.Error(),
			})
		default:
			log.Error(err, nil)
			return c.Status(fiber.StatusInternalServerError).JSON(response{
				Status: errors.ErrInternalServerError.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(response{
		Status: errors.Okay,
	})
}

// ResendEmailConfirmation is a function that is used to resend the email confirmation to the user
func (Email) ResendEmailConfirmation(c *fiber.Ctx, h *initialize.H, env *config.Env) error {
	userID := c.Locals(config.Enums{}.USER()).(string)

	err := utils.Email{}.ResendConfirmaton(h, env, userID)
	if err != nil {
		if err == errors.ErrUnauthorized {
			return c.Status(fiber.StatusBadRequest).JSON(response{
				Status: err.Error(),
			})
		}

		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response{
		Status: errors.Okay,
	})
}
