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
