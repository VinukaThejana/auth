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

// Login is a function that allows the user to login with the email and password
func (Auth) Login(c *fiber.Ctx, h *initialize.H, env *config.Env) error {
	var payload *schemas.LoginInput
	if err := c.BodyParser(&payload); err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	if err := payload.Vaidate(); err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	var user models.User
	result := h.DB.DB.First(&user, "email = ?", payload.Email)
	if result.Error != nil {
		if result.Error != gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusBadRequest).JSON(response{
				Status: errors.ErrBadRequest.Error(),
			})
		}

		log.Error(result.Error, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Status: errors.ErrUnauthorized.Error(),
		})
	}

	accessTokenDetails, err := utils.Token{}.CreateToken(h, schemas.User{
		ID:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
	}, env.AccessTokenPrivateKey, env.AccessTokenExpires)
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	refreshTokenDetails, err := utils.Token{}.CreateToken(h, schemas.User{
		ID:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
	}, env.RefreshTokenPrivateKey, env.RefreshTokenExpires)
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    *accessTokenDetails.Token,
		Path:     "/",
		MaxAge:   env.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: true,
		Domain:   "localhost",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    *refreshTokenDetails.Token,
		Path:     "/",
		MaxAge:   env.RefreshTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: true,
		Domain:   "localhost",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "logged_in",
		Value:    "true",
		Path:     "/",
		MaxAge:   env.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: false,
		Domain:   "localhost",
	})

	return c.Status(fiber.StatusOK).JSON(response{
		Status: errors.Okay,
	})
}

// RefreshToken is a function that is used to refresh the token
func (Auth) RefreshToken(c *fiber.Ctx, h *initialize.H, env *config.Env) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	tokenClaims, err := utils.Token{}.ValidateToken(h, refreshToken, env.RefreshTokenPublicKey)
	if err != nil {
		if err == errors.ErrUnauthorized {
			return c.Status(fiber.StatusUnauthorized).JSON(response{
				Status: err.Error(),
			})
		}

		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	accessTokenDetails, err := utils.Token{}.CreateToken(h, tokenClaims.User, env.AccessTokenPrivateKey, env.AccessTokenExpires)
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    *accessTokenDetails.Token,
		Path:     "/",
		MaxAge:   env.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: true,
		Domain:   "localhost",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "logged_in",
		Value:    "true",
		Path:     "/",
		MaxAge:   env.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: false,
		Domain:   "localhost",
	})

	return c.Status(fiber.StatusOK).JSON(response{
		Status: errors.Okay,
	})
}

func (Auth) Logout(c *fiber.Ctx, h *initialize.H, env *config.Env) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Status: errors.ErrUnauthorized.Error(),
		})
	}

	_, err := utils.Token{}.ValidateToken(h, refreshToken, env.RefreshTokenPublicKey)
	if err != nil {
		if err == errors.ErrUnauthorized {
			return c.Status(fiber.StatusUnauthorized).JSON(response{
				Status: errors.ErrUnauthorized.Error(),
			})
		}

		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	ctx := context.TODO()
	err = utils.Token{}.DeleteToken(h, refreshToken)
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}
	h.R.RS.Del(ctx, c.Locals(config.Enums{}.ACCESSTOKENUUID()).(string))

	expired := time.Now().Add(-time.Hour * 24)
	c.Cookie(&fiber.Cookie{
		Name:    "access_token",
		Value:   "",
		Expires: expired,
	})

	c.Cookie(&fiber.Cookie{
		Name:    "refresh_token",
		Value:   "",
		Expires: expired,
	})

	c.Cookie(&fiber.Cookie{
		Name:    "logged_in",
		Value:   "",
		Expires: expired,
	})

	return c.Status(fiber.StatusOK).JSON(response{
		Status: errors.Okay,
	})
}
