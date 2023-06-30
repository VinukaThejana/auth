package models

import (
	"time"

	"github.com/google/uuid"
)

// Sessions is a model that represents the sessions in the relational database
type Sessions struct {
	TokenID   uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	IPAddress string
	Location  string
	Device    string
	OS        string
	LoginAt   time.Time `gorm:"not null;default:now()"`
	ExpiresAt int64     `gorm:"not null"`
}
