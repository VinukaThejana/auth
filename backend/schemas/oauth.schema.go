package schemas

// BasicOAuthProvider contains all the common feils related to oauth providers
type BasicOAuthProvider struct {
	ID       string
	Name     string
	Username string
	Email    *string
}

// GitHub struct contains the needed data that is received from GitHub after OAuth login
type GitHub struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Username  string  `json:"login"`
	AvatarURL string  `json:"avatar_url"`
	Email     *string `json:"email"`
}

// GetEmailFromPayload is a function that is used to extract the email feild from the api
func (GitHub) GetEmailFromPayload(payload map[string]interface{}) *string {
	if email, ok := payload["email"].(*string); ok && email != nil {
		return email
	}

	return nil
}
