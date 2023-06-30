package models

import (
	"time"

	"github.com/google/uuid"
)

// Sessions is a model that represents the sessions tabel in the database this table is used to
// store session related data in the database
type Sessions struct {
	UserID    *uuid.UUID `gorm:"type:uuid;primary_key"`
	TokenID   *uuid.UUID `gorm:"type:uuid;primary_key"`
	ExpiresAt time.Time  `gorm:"not null"`
}
