package initialize

import (
	"context"

	"github.com/VinukaThejana/auth/backend/config"
	"github.com/redis/go-redis/v9"
)

// Redis is a struct that contains the Redis database instances
type Redis struct {
	RR *redis.Client
	RS *redis.Client
	RE *redis.Client
}

func connnect(addrr string) *redis.Client {
	r := redis.NewClient(&redis.Options{
		Addr: addrr,
	})

	if err := r.Ping(context.Background()); err.Err() != nil {
		log.Errorf(err.Err(), nil)
	}

	return r
}

// InitiRedis is a function that is used to intiialize the Redis database
// instance
func (h *H) InitiRedis(env *config.Env) {
	h.R = &Redis{
		RS: connnect(env.RedisSessionURL),
		RR: connnect(env.RedisRatelimiterURL),
		RE: connnect(env.RedisEmailURL),
	}
}
