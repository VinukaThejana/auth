package services

import (
	"fmt"

	"github.com/VinukaThejana/auth/backend/errors"
	"github.com/VinukaThejana/auth/backend/initialize"
	"github.com/VinukaThejana/auth/backend/models"
	"github.com/VinukaThejana/auth/backend/schemas"
	"gorm.io/gorm"
)

// GitHub contains all GitHub related OAuth operations
type GitHub struct{}

func create(h *initialize.H, profile schemas.BasicOAuthProvider, provider string) (newUser models.User, err error) {
	verified := true

	newUser.Name = profile.Name
	newUser.Username = profile.Username
	newUser.Verified = &verified
	newUser.Provider = &provider
	newUser.ProviderID = profile.ID

	if profile.Email != nil {
		newUser.Email = *profile.Email
	}

	newUser, err = User{}.Create(h, newUser)
	if err != nil {
		return models.User{}, err
	}

	return newUser, nil
}

// GitHubOAuth is a function to login / register users with GitHub accounts
func (GitHub) GitHubOAuth(h *initialize.H, profile schemas.GitHub) (user models.User, err error) {
	provider := models.GitHubProvider

	err = h.DB.DB.Where("provider = ?", models.GitHubProvider).Where("provider_id = ?", fmt.Sprint(profile.ID)).First(&user).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return models.User{}, err
		}

		if profile.Email == nil {
			ok, err := User{}.IsUsernameAvailable(h, profile.Username)
			if err != nil {
				return models.User{}, err
			}

			if !ok {
				// INFO: Prompt the user to choose the username
				return models.User{}, errors.ErrAddAUsername
			}

			user, err = create(h, schemas.BasicOAuthProvider{
				ID:       fmt.Sprint(profile.ID),
				Name:     profile.Name,
				Username: profile.Username,
				Email:    profile.Email,
			}, provider)
			if err != nil {
				return models.User{}, err
			}

			return user, nil
		}

		id, ok, verified, err := User{}.IsEmailAvailable(h, *profile.Email)
		if err != nil {
			return models.User{}, err
		}

		if !ok && verified {
			err := h.DB.DB.Save(&models.User{
				ID:         id,
				Provider:   &provider,
				ProviderID: fmt.Sprint(profile.ID),
			}).Error
			if err != nil {
				return models.User{}, nil
			}

			user.ID = id
			err = h.DB.DB.First(&user).Error
			if err != nil {
				return models.User{}, err
			}

			return user, nil
		}

		if !ok {
			// TODO: Think of the way to handle the scenario where the user email address is available in the database but
			// not yet verified
			return models.User{}, fmt.Errorf("FIX ME")
		}

		user, err = create(h, schemas.BasicOAuthProvider{
			ID:       fmt.Sprint(profile.ID),
			Name:     profile.Name,
			Username: profile.Username,
			Email:    profile.Email,
		}, provider)
		if err != nil {
			return models.User{}, err
		}

		return user, nil
	}

	return user, nil
}
