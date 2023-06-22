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

	auth controllers.Auth
	user controllers.User
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

  authG := app.Group("/auth")
  authG.Post("/register", func(c *fiber.Ctx) error {
    return auth.Register(c, &h);
  })
  authG.Post("/login", func(c *fiber.Ctx) error {
    return auth.Login(c, &h, &env)
  })
  authG.Post("/refresh", func(c *fiber.Ctx) error {
    return auth.RefreshToken(c, &h, &env)
  })
  authG.Post("/refresh", func(c *fiber.Ctx) error {
    return auth.RefreshToken(c, &h, &env)
  })

  userG := app.Group("/user", func(c *fiber.Ctx) error {
    return middleware.CheckAuth(c, &h, &env)
  })
  userG.Get("/", func(c *fiber.Ctx) error {
    return user.GetUser(c, &h)
  })

  log.Errorf(app.Listen(fmt.Sprintf(":%s", env.Port)), nil)
}
