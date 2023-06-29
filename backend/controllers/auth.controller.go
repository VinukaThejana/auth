package controllers

import (
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
func (Auth) Register(c *fiber.Ctx, h *initialize.H, env *config.Env) error {
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
		Username: payload.Username,
		Email:    payload.Email,
		Password: string(hashedPassword),
	}

	result := h.DB.DB.Create(&newUser)
	if result.Error != nil {
		if ok := (errors.CheckDBError{}.DuplicateKey(result.Error)); ok {
			return c.Status(fiber.StatusBadRequest).JSON(response{
				Status: errors.ErrBadRequest.Error(),
			})
		}

		log.Error(result.Error, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	go utils.Email{}.SendConfirmation(h, env, newUser.Email, newUser.ID.String())

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

	if err := payload.Validate(); err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Status: errors.ErrBadRequest.Error(),
		})
	}

	var user models.User
	if payload.Username != "" {
		result := h.DB.DB.First(&user, "username = ?", payload.Username)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusBadRequest).JSON(response{
					Status: errors.ErrIncorrectCredentials.Error(),
				})
			}

			log.Error(result.Error, nil)
			return c.Status(fiber.StatusInternalServerError).JSON(response{
				Status: errors.ErrInternalServerError.Error(),
			})
		}
	} else {
		result := h.DB.DB.First(&user, "email = ?", payload.Email)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusBadRequest).JSON(response{
					Status: errors.ErrIncorrectCredentials.Error(),
				})
			}

			log.Error(result.Error, nil)
			return c.Status(fiber.StatusInternalServerError).JSON(response{
				Status: errors.ErrInternalServerError.Error(),
			})
		}
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Status: errors.ErrUnauthorized.Error(),
		})
	}

	accessTokenDetails, err := utils.Token{}.CreateAccessToken(h, user.ID.String(), env.AccessTokenPrivateKey, env.AccessTokenExpires)
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	refreshTokenDetails, err := utils.Token{}.CreateRefreshToken(h, user.ID.String(), env.RefreshTokenPrivateKey, env.RefreshTokenExpires, struct {
		IPAddress       string
		Location        string
		Device          string
		OS              string
		AccessTokenUUID string
	}{
		AccessTokenUUID: accessTokenDetails.TokenUUID,
	})
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

	tokenClaims, _, err := utils.Token{}.ValidateRefreshToken(h, refreshToken, env.RefreshTokenPublicKey)
	if err != nil {
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

	accessTokenDetails, err := utils.Token{}.CreateAccessToken(h, tokenClaims.UserID, env.AccessTokenPrivateKey, env.AccessTokenExpires)
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

// Logout is a function that is used to logout the user
func (Auth) Logout(c *fiber.Ctx, h *initialize.H, env *config.Env) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Status: errors.ErrUnauthorized.Error(),
		})
	}

	tokenDetails, tokenValue, err := utils.Token{}.ValidateRefreshToken(h, refreshToken, env.RefreshTokenPublicKey)
	log.Error(err, nil)
	if err != nil {
		if err == errors.ErrUnauthorized {
			return c.Status(fiber.StatusUnauthorized).JSON(response{
				Status: errors.ErrUnauthorized.Error(),
			})
		}

		if ok := (errors.CheckTokenError{}.Expired(err)); ok {
			return c.Status(fiber.StatusUnauthorized).JSON(response{
				Status: errors.ErrAccessTokenExpired.Error(),
			})
		}

		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	err = utils.Token{}.DeleteToken(h, *&tokenDetails.TokenUUID, *&tokenValue.AccessTokenUUID)
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

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

// CheckUsername is a function that is used to check wether the provided username is available
// in the platform
func (Auth) CheckUsername(c *fiber.Ctx, h *initialize.H) error {
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
