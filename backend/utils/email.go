package utils

import (
	"context"
	"time"

	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/templates"
	"github.com/google/uuid"
	"github.com/resendlabs/resend-go"
)

var (
	resendEmailFrom    = "onboarding@resend.dev"
	resendReplyToEmail = "onboarding@resend.dev"
)

// Email is a struct that contains email related functionality
type Email struct{}

// SendConfirmation is a function that is used to send a confirmation email to the client
func (Email) SendConfirmation(h *initialize.H, env *config.Env, email, userID string) {
	token := uuid.New()
	ctx := context.TODO()
	h.R.RE.SetNX(ctx, token.String(), userID, 30*60*time.Second)

	emailTemplate, err := templates.Email{}.GetEmailConfirmationTmpl(token.String())
	if err != nil {
		log.Error(err, nil)
	}

	client := resend.NewClient(env.ResendAPIKey)
	params := &resend.SendEmailRequest{
		From:    resendEmailFrom,
		To:      []string{"vinuka.t@pm.me"},
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
