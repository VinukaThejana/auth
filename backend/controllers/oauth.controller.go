package controllers

import (
	"fmt"
	"net/url"

	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/errors"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/schemas"
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

	return c.Status(fiber.StatusOK).JSON(schemas.FilterUserRecord(&user))
}
