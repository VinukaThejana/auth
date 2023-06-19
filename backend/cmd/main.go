// Authentication backend to register, login and validate users
package main

import (
	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/go-utils/logger"
)

var (
	log logger.Logger
	env config.Env
	h   initialize.H
)

func init() {
	env.Load()

	h.InitDB(&env)
	h.InitiRedis(&env)
}

func main() {
}
