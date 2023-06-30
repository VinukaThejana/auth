package initialize

import (
	"fmt"

	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/models"
	"github.com/fatih/color"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// DB contains the struct to hold the database instance
type DB struct {
	DB *gorm.DB
}

// InitDB initializes a connnection to the postgress database
func (h *H) InitDB(env *config.Env) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Colombo", env.DBHost, env.DBUser, env.DBPassword, env.DBName, fmt.Sprint(env.DBPort))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Errorf(err, nil)
	}

	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	db.Logger = gormLogger.Default.LogMode(gormLogger.Info)

	color.Blue("Running migrations ... ")
	err = db.AutoMigrate(models.User{}, models.Sessions{})
	if err != nil {
		errMsg := "Error running migrations !"
		log.Errorf(err, &errMsg)
	}

	h.DB = &DB{
		DB: db,
	}
}
