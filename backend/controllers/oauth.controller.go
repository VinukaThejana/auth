package controllers

import (
	"fmt"
	"net/url"

	"github.com/VinukaThejana/auth/backend/config"
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
