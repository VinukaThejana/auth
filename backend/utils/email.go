package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/errors"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/models"
	"github.com/VinukaThejana/auth/backend/templates"
	"github.com/google/uuid"
	"github.com/resendlabs/resend-go"
	"gorm.io/gorm"
)

var (
	resendEmailFrom                 = "onboarding@resend.dev"
	resendReplyToEmail              = "onboarding@resend.dev"
	emailConfirmationExpirationTime = 30 * 60 * time.Second
)

// Email is a struct that contains email related functionality
type Email struct{}

// SendConfirmation is a function that is used to send a confirmation email to the client
func (Email) SendConfirmation(h *initialize.H, env *config.Env, email, userID string) {
	token := uuid.New()
	ctx := context.TODO()
	h.R.RE.SetNX(ctx, token.String(), fmt.Sprintf("%s+%s", userID, email), emailConfirmationExpirationTime)

	emailTemplate, err := templates.Email{}.GetEmailConfirmationTmpl(token.String())
	if err != nil {
		log.Error(err, nil)
	}

	client := resend.NewClient(env.ResendAPIKey)
	params := &resend.SendEmailRequest{
		From:    resendEmailFrom,
		To:      []string{email},
		Html:    emailTemplate,
		Subject: "Email confirmation",
		ReplyTo: resendReplyToEmail,
	}
	send, err := client.Emails.Send(params)
	if err != nil {
		log.Error(err, nil)
	}
	log.Success(send.Id)
}

// ConfirmEmail is a function that is used to confirm the email of the user with the provided token
func (Email) ConfirmEmail(h *initialize.H, token, userID string) error {
	var user struct {
		ID    string
		Email string
	}

	_, err := uuid.Parse(token)
	if err != nil {
		return errors.ErrBadRequest
	}

	ctx := context.TODO()
	val := h.R.RE.Get(ctx, token).Val()
	if val == "" {
		return errors.ErrEmailConfirmationExpired
	}

	var found bool
	user.ID, user.Email, found = strings.Cut(val, "+")
	if !found {
		return errors.ErrUnauthorized
	}

	if user.ID != userID {
		return errors.ErrUnauthorized
	}

	err = h.DB.DB.Model(&models.User{}).Where("id = ?", userID).Where("email = ?", user.Email).Update("verified", true).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUnauthorized
		}

		return err
	}

	h.R.RE.Del(ctx, token)
	return nil
}

// ResendConfirmaton is a function that is used to resend the email confirmation
func (Email) ResendConfirmaton(h *initialize.H, env *config.Env, userID string) error {
	var user models.User
	err := h.DB.DB.First(&user, "id = ?", userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUnauthorized
		}

		return err
	}

	Email{}.SendConfirmation(h, env, user.Email, userID)
	return nil
}
