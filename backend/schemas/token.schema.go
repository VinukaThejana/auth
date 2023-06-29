package schemas

import "time"

// RefreshTokenDetails is a struct that contains details about the refresh token
type RefreshTokenDetails struct {
	UserID          string
	IPAddress       string
	Location        string
	Device          string
	OS              string
	LoginAt         time.Time
	AccessTokenUUID string
}
