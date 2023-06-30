package models

import (
	"time"

	"github.com/google/uuid"
)

// User struct reprents the user table in the relational database
type User struct {
	ID        *uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name      string     `gorm:"type:varchar(100);not null"`
	Username  string     `gorm:"type:varchar(100);uniqueIndex;not null"`
	Email     string     `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string     `gorm:"type:varchar(100);not null"`
	Role      *string    `gorm:"type:varchar(50);default:'user';not null"`
	Provider  *string    `gorm:"type:varchar(50);default:'local';not null"`
	Verified  *bool      `gorm:"not null;default:false"`
	CreatedAt *time.Time `gorm:"not null;default:now()"`
	UpdatedAt *time.Time `gorm:"not null;default:now()"`
	Sessions  []Sessions `gorm:"foreignKey:UserID"`
}
