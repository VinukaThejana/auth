package initialize

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/VinukaThejana/auth/backend/config"
	"github.com/gofiber/storage/redis"
)

// Storage is struct that contains the connections to the Redis storage for ratelimiting
type Storage struct {
	S *redis.Storage
}

// InitStorage is a function that is used to initialize the Redis ratelimiter
// storage only
func (h *H) InitStorage(env *config.Env) {
	data := strings.Split(env.RedisRatelimiterURL, ":")
	if len(data) != 2 {
		log.Errorf(fmt.Errorf("Invalid Redis URL"), nil)
	}
	host := data[0]
	port, err := strconv.Atoi(data[1])
	if err != nil {
		log.Errorf(err, nil)
	}

	store := redis.New(redis.Config{
		Host:     host,
		Port:     port,
		Username: "",
	})

	h.S = &Storage{
		S: store,
	}
}
