package schemas

import (
	"fmt"
	"time"

	"github.com/VinukaThejana/auth/backend/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// RegisterInput is a struct that defines what the server expects from the user upon registration
type RegisterInput struct {
	Name                 string `json:"name" validate:"required,min=3,max=40"`
	Username             string `json:"username" validate:"required,min=3,max=20"`
	Email                string `json:"email" validate:"required,email"`
	Password             string `json:"password" validate:"required,min=8,max=200"`
	PasswordConfirmation string `json:"password_confirm" validate:"required,min=8,max=200"`
}

// Validate is a function that is used to vaidate user input upon registration
func (rI RegisterInput) Validate() (err error) {
	v := validator.New()
	err = v.Struct(rI)
	return err
}

// LoginInput is a struct that defines what the server expects from the user upon login
type LoginInput struct {
	Email    string `json:"email" validate:"omitempty,email"`
	Username string `json:"username" validate:"omitempty,min=3,max=20"`
	Password string `json:"password" validate:"required,min=8,max=200"`
}

// Validate is a function that is used to vaidate user input upon login
func (lI LoginInput) Validate() (err error) {
	v := validator.New()
	err = v.Struct(lI)
	if lI.Email == "" && lI.Username == "" {
		err = fmt.Errorf("username and email both cannot be empty")
	}

	return err
}

// User struct contians the most basic data that needs to be stored from a user
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// UserResponse is a struct that contains all the relevant feilds of the models.User when sending the
// user session to the client side
type UserResponse struct {
	ID        uuid.UUID `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Username  string    `json:"username,omitempty"`
	Email     string    `json:"email,omitempty"`
	Role      string    `json:"role,omitempty"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FilterUserRecord is a funcion that is used to filter the models.User struct to a client freindly manner
func FilterUserRecord(user *models.User) UserResponse {
	return UserResponse{
		ID:        *user.ID,
		Name:      user.Name,
		Username:  user.Username,
		Email:     user.Email,
		Role:      *user.Role,
		Provider:  *user.Provider,
		CreatedAt: *user.CreatedAt,
		UpdatedAt: *user.UpdatedAt,
	}
}
