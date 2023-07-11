// Authentication backend to register, login and validate users
package main

import (
	"fmt"
	"time"

	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/controllers"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/middleware"
	"github.com/VinukaThejana/go-utils/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
)

var (
	log logger.Logger
	env config.Env
	h   initialize.H

	auth  controllers.Auth
	user  controllers.User
	email controllers.Email
	oauth controllers.OAuth
)

func init() {
	env.Load()

	h.InitDB(&env)
	h.InitiRedis(&env)
	h.InitStorage(&env)
}

func main() {
	app := fiber.New()

	app.Use(fiberLogger.New())
	app.Use(cors.New(cors.Config{
		AllowHeaders:     "*",
		AllowOrigins:     "*",
		AllowCredentials: true,
		AllowMethods:     "*",
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		LimiterMiddleware:      limiter.SlidingWindow{},
		Storage:                h.S.S,
	}))
	app.Get("/metrics", monitor.New(monitor.Config{
		Title: "auth",
	}))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("https://app.theneo.io/szeeta/auth")
	})

	authG := app.Group("/auth")
	authG.Post("/register", func(c *fiber.Ctx) error {
		return auth.Register(c, &h, &env)
	})
	authG.Post("/login", func(c *fiber.Ctx) error {
		return auth.Login(c, &h, &env)
	})
	authG.Post("/refresh", func(c *fiber.Ctx) error {
		return auth.RefreshToken(c, &h, &env)
	})
	authG.Post("/validate/username", func(c *fiber.Ctx) error {
		return auth.CheckUsername(c, &h)
	})
	authG.Post("/logout", func(c *fiber.Ctx) error {
		return auth.Logout(c, &h, &env)
	})

	oauthG := app.Group("/oauth")
	oauthG.Route("/redirects", func(router fiber.Router) {
		router.Get("/github", func(c *fiber.Ctx) error {
			return oauth.RedirectToGitHubOAuthFlow(c, &env)
		})
	})
	oauthG.Route("/sessions", func(router fiber.Router) {
		router.Get("/github", func(c *fiber.Ctx) error {
			return oauth.GithubOAuthCallback(c, &h, &env)
		})
	})

	userG := app.Group("/user", func(c *fiber.Ctx) error {
		return middleware.CheckAuth(c, &h, &env)
	})
	userG.Get("/", func(c *fiber.Ctx) error {
		return user.GetUser(c, &h)
	})
	userG.Route("/update", func(router fiber.Router) {
		router.Post("/email", func(c *fiber.Ctx) error {
			return user.UpdateEmail(c, &h, &env)
		})
		router.Post("/username", func(c *fiber.Ctx) error {
			return user.UpdateUsername(c, &h)
		})
		router.Post("/name", func(c *fiber.Ctx) error {
			return user.UpdateName(c, &h)
		})
	})
	userG.Route("/auth", func(router fiber.Router) {
		router.Get("/devices", func(c *fiber.Ctx) error {
			return user.GetAuthInstances(c, &h, &env)
		})
		router.Post("/confirm", func(c *fiber.Ctx) error {
			return user.ConfirmAction(c, &h, &env)
		})
		router.Post("/logout-from-device", func(c *fiber.Ctx) error {
			return user.LogoutFromDevice(c, &h)
		})
	})

	emailG := app.Group("/email", func(c *fiber.Ctx) error {
		return middleware.CheckAuth(c, &h, &env)
	})
	emailG.Route("/confirmation", func(router fiber.Router) {
		router.Get("/", func(c *fiber.Ctx) error {
			return email.ConfirmEmail(c, &h, &env)
		})
		router.Get("/resend", func(c *fiber.Ctx) error {
			return email.ResendEmailConfirmation(c, &h, &env)
		})
	})

	log.Errorf(app.Listen(fmt.Sprintf(":%s", env.Port)), nil)
}
