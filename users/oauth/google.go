package oauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	scope       = "https://www.googleapis.com/auth/userinfo.profile"
	userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
)

type Config struct {
	*oauth2.Config
	State string
}

type User struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

func NewConfig(clientId, clientSecret, redirectURL string) Config {
	oauthCon := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{scope},
		Endpoint:     google.Endpoint,
	}

	return Config{Config: oauthCon}
}

func (conf *Config) Profile(ctx context.Context, code string) (User, *oauth2.Token, error) {
	token, err := conf.Config.Exchange(ctx, code)
	if err != nil {
		return User{}, &oauth2.Token{}, err
	}

	resp, err := http.Get(userInfoURL + url.QueryEscape(token.AccessToken))
	if err != nil {
		return User{}, token, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return User{}, token, err
	}

	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return User{}, token, err
	}

	return user, token, nil
}
