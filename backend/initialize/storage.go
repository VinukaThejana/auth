package initialize

import (
	"github.com/VinukaThejana/auth/backend/config"
	"github.com/gofiber/storage/redis"
)

type Storage struct {
	S *redis.Storage
}

// InitStorage is a function that is used to initialize the Redis ratelimiter
// storage only
func (h *H) InitStorage(env *config.Env) {
	store := redis.New(redis.Config{
    URL: env.RedisRatelimiterURL,
    Reset: false,
  })

  h.S = &Storage{
    S: store,
  }
}
