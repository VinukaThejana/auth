package middleware

import (
	"strings"

	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/errors"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/utils"
	"github.com/gofiber/fiber/v2"
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
		if err == errors.ErrUnauthorized {
			return c.Status(fiber.StatusUnauthorized).JSON(response{
				Status: err.Error(),
			})
		}

		if ok := (errors.CheckTokenError{}.Expired(err)); ok {
			return c.Status(fiber.StatusUnauthorized).JSON(response{
				Status: errors.ErrAccessTokenExpired.Error(),
			})
		}

		log.Error(err, nil)
		return c.Status(fiber.StatusForbidden).JSON(response{
			Status: errors.ErrAccessTokenNotProvided.Error(),
		})
	}

	c.Locals(config.Enums{}.USER(), tokenClaims.UserID)
	c.Locals(config.Enums{}.ACCESSTOKENUUID(), tokenClaims.TokenUUID)

	return c.Next()
}
