package controllers

import (
	"fmt"
	"net/url"

	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/errors"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/services"
	"github.com/VinukaThejana/auth/backend/utils"
	"github.com/gofiber/fiber/v2"
)

// OAuth related controllers
type OAuth struct{}

// RedirectToGitHubOAuthFlow controller redirects to the github oauth login page
func (OAuth) RedirectToGitHubOAuthFlow(c *fiber.Ctx, env *config.Env) error {
	options := url.Values{
		"client_id":    []string{env.GithubClientID},
		"redirect_uri": []string{env.GithubRedirectURL},
		"scope":        []string{"user:email"},
		"state":        []string{env.GithubFromURL},
	}

	githubRedirectURL := fmt.Sprintf("%s?%s", env.GithubRootURL, options.Encode())
	return c.Redirect(githubRedirectURL)
}

// GithubOAuthCallback is a function that is used to continue the flow with github once the user
// authorized the Github account
func (OAuth) GithubOAuthCallback(c *fiber.Ctx, h *initialize.H, env *config.Env) error {
	code := c.Query("code")
	// state := c.Query("state")

	if code == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Status: errors.ErrUnauthorized.Error(),
		})
	}

	accessToken, err := utils.OAuth{}.GetGitHubAccessToken(code, env)
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	userDetails, err := utils.OAuth{}.GetGitHubUser(*accessToken)
	if err != nil {
		log.Error(err, nil)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	fmt.Println(*userDetails, userDetails.ID)

	user, err := services.GitHub{}.GitHubOAuth(h, *userDetails)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Status: errors.ErrInternalServerError.Error(),
		})
	}

	go func() {
		utils.Token{}.DeleteExpiredTokens(h, user.ID.String())
	}()

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
