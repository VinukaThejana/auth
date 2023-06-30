package controllers

import (
	"context"
	"encoding/json"
	"time"

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

// GetAuthInstances is a function that is used to obtain the authed instances of the user
func (User) GetAuthInstances(c *fiber.Ctx, h *initialize.H, env *config.Env) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	tokenClaims, _, err := utils.Token{}.ValidateRefreshToken(h, refreshToken, env.RefreshTokenPublicKey)
	if err != nil {
		log.Error(err, nil)
		if err == errors.ErrUnauthorized {
			return c.Status(fiber.StatusUnauthorized).JSON(response{
				Status: err.Error(),
			})
		}

		if ok := (errors.CheckTokenError{}.Expired(err)); ok {
			return c.Status(fiber.StatusUnauthorized).JSON(response{
				Status: errors.ErrRefreshTokenExpired.Error(),
			})
		}

		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	var sessions []models.Sessions
	err = h.DB.DB.Where("user_id = ? AND expires_at > ?", tokenClaims.UserID, time.Now().UTC().Unix()).Find(&sessions).Error
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(sessions)
}

// LogoutFromDevice is a function that is used to logut a user from a logged in device
func (User) LogoutFromDevice(c *fiber.Ctx, h *initialize.H) error {
	userID := c.Locals(config.Enums{}.USER()).(string)
	accessTokenUUID := c.Locals(config.Enums{}.ACCESSTOKENUUID()).(string)

	var payload struct {
		RefreshTokenUUID string `json:"refresh_token_uuid"`
	}
	if err := c.BodyParser(&payload); err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	ctx := context.TODO()
	val := h.R.RS.Get(ctx, payload.RefreshTokenUUID).Val()
	if val == "" {
		go func() {
			utils.Token{}.DeleteExpiredTokens(h, userID)
		}()
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	var tokenValue schemas.RefreshTokenDetails
	err := json.Unmarshal([]byte(val), &tokenValue)
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	if tokenValue.UserID != userID {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Status: errors.ErrUnauthorized.Error(),
		})
	}

	err = utils.Token{}.DeleteToken(h, payload.RefreshTokenUUID, tokenValue.AccessTokenUUID)
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	if tokenValue.AccessTokenUUID == accessTokenUUID {
		// TODO: Do some crazy redirect as the currently logged in session is deleted
	}

	return c.Status(fiber.StatusOK).JSON(response{
		Status: errors.Okay,
	})
}
