package middleware

import (
	"fmt"
	"strings"

	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/errors"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/models"
	"github.com/VinukaThejana/auth/backend/schemas"
	"github.com/VinukaThejana/auth/backend/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CheckAuth is a middleware function that is used to check wether the user is authed
func CheckAuth(c *fiber.Ctx, h *initialize.H, env *config.Env) error {
	var accessToken string
	authorization := c.Get("Authorization")

	if strings.HasPrefix(authorization, "Bearer ") {
		accessToken = strings.TrimPrefix(authorization, "Bearer ")
	} else {
		if c.Cookies("access_token") != "" {
			accessToken = c.Cookies("accessToken")
		} else {
			return c.Status(fiber.StatusForbidden).JSON(response{
				Status: errors.ErrAccessTokenNotProvided.Error(),
			})
		}
	}

	if accessToken == "" {
		return c.Status(fiber.StatusForbidden).JSON(response{
			Status: errors.ErrAccessTokenNotProvided.Error(),
		})
	}

	tokenClaims, err := utils.Token{}.ValidateToken(h, accessToken, env.AccessTokenPublicKey)
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusForbidden).JSON(response{
			Status: errors.ErrAccessTokenNotProvided.Error(),
		})
	}

	var user models.User
	err = h.DB.DB.First(&user, "id = ?", tokenClaims.UserID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Error(fmt.Errorf("user with the uid = %s does not exist in the database", tokenClaims.UserID), nil)
			return c.Status(fiber.StatusUnauthorized).JSON(response{
				Status: errors.ErrUnauthorized.Error(),
			})
		}

		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	c.Locals("user", schemas.FilterUserRecord(&user))
	c.Locals("access_token_uuid", tokenClaims.TokenUUID)

	return c.Next()
}
