package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/VinukaThejana/auth/backend/config"
	"github.com/VinukaThejana/auth/backend/schemas"
)

// OAuth related utilities
type OAuth struct{}

// GetGitHubAccessToken is a function that is used to get the access token from GitHub
func (OAuth) GetGitHubAccessToken(code string, env *config.Env) (accessToken *string, err error) {
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	query := url.Values{
		"code":          []string{code},
		"client_id":     []string{env.GithubClientID},
		"client_secret": []string{env.GithubClientSecret},
	}.Encode()
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://github.com/login/oauth/access_token?%s", bytes.NewBufferString(query)), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Could not retrieve the access token")
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	parsedQuery, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}
	if len(parsedQuery["access_token"]) == 0 {
		err = fmt.Errorf("Access token is not provided")
		return nil, err
	}

	token := parsedQuery["access_token"][0]
	return &token, nil
}

// GetGitHubUser is a fucntion to get the GitHub user from the access token provided from github
func (OAuth) GetGitHubUser(accessToken string) (*schemas.GitHub, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to fetch user data from GitHub")
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err = json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	return &schemas.GitHub{
		ID:        payload["id"].(float64),
		Name:      payload["name"].(string),
		Username:  payload["login"].(string),
		AvatarURL: payload["avatar_url"].(string),
		Email:     schemas.GitHub{}.GetEmailFromPayload(payload),
	}, nil
}
