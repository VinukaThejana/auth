package config

import (
	"time"

	"github.com/spf13/viper"
)

type Env struct {
	DBHost     string `mapstructure:"POSTGRES_HOST" validate:"required"`
	DBPort     int    `mapstructure:"POSTGRES_PORT" validate:"required,min=1,max=65535"`
	DBUser     string `mapstructure:"POSTGRES_USER" validate:"required,min=3,max=15"`
	DBPassword string `mapstructure:"POSTGRES_PASSWORD" validate:"required"`
	DBName     string `mapstructure:"POSTGRES_DB" validate:"required"`

	DSN string `mapstructure:"DATABASE_URL" validate:"required"`

	RedisURL string `mapstructure:"REDIS_URL" validate:"required"`

	Port string `mapstructure:"PORT" validate:"required"`

	AccessTokenPrivateKey string        `mapstructure:"ACCESS_TOKEN_PRIVATE_KEY" validate:"required"`
	AccessTokenPublicKey  string        `mapstructure:"ACCESS_TOKEN_PUBLIC_KEY" validate:"required"`
	AccessTokenExpires    time.Duration `mapstructure:"ACCESS_TOKEN_EXPIRED_IN" validate:"required"`
	AccessTokenMaxAge     int           `mapstructure:"ACCESS_TOKEN_MAXAGE" validate:"required"`

	RefreshTokenPrivateKey string        `mapstructure:"REFRESH_TOKEN_PRIVATE_KEY" validate:"required"`
	RefreshTokenPublicKey  string        `mapstructure:"REFRESH_TOKEN_PUBLIC_KEY" validate:"required"`
	RefreshTokenExpires    time.Duration `mapstructure:"REFRESH_TOKEN_EXPIRED_IN" validate:"required"`
	RefreshTokenMaxAge     int           `mapstructure:"REFRESH_TOKEN_MAXAGE" validate:"required"`
}

// Load is a function that is used to load the env variables from the env file
// to the env struct (Env)
func (e *Env) Load() {
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")

	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Errorf(err, nil)
	}

	err = viper.Unmarshal(&e)
	if err != nil {
		log.Errorf(err, nil)
	}

	log.Validatef(e)
}
