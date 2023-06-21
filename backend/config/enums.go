package config

// Enums contains needed enums
type Enums struct{}

// USER contains the user enum
func (Enums) USER() string {
	return "user"
}

// ACCESSTOKENUUID contains the access token uuid enum
func (Enums) ACCESSTOKENUUID() string {
	return "access_token_uuid"
}
