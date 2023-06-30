package models

import (
	"github.com/google/uuid"
)

// Sessions is a model that represents the sessions tabel in the database this table is used to
// store session related data in the database
type Sessions struct {
	TokenID   uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	ExpiresAt int64     `gorm:"not null"`
}
