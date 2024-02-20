// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package oauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	mfclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/users"
	"golang.org/x/oauth2"
)

const (
	userInfoEndpoint = "/userinfo?access_token="
	authEndpoint     = "/oauth2/auth"
	TokenEndpoint    = "/oauth2/token"
)

var scopes = []string{
	"email",
	"profile",
	"offline_access",
}

type Config struct {
	config        *oauth2.Config
	state         string
	baseURL       string
	uiRedirectURL string
	userInfoURL   string
}

func NewConfig(baseURL, clientID, clientSecret, state, redirectURL, uiRedirectURL string) Config {
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  baseURL + authEndpoint,
			TokenURL: baseURL + TokenEndpoint,
		},
		RedirectURL: redirectURL,
		Scopes:      scopes,
	}

	return Config{
		config:        cfg,
		state:         state,
		baseURL:       baseURL,
		uiRedirectURL: uiRedirectURL,
		userInfoURL:   baseURL + userInfoEndpoint,
	}
}

func (cfg *Config) Profile(ctx context.Context, code string) (mfclients.Client, *oauth2.Token, error) {
	token, err := cfg.config.Exchange(ctx, code)
	if err != nil {
		return mfclients.Client{}, &oauth2.Token{}, err
	}

	client, err := cfg.Client(token.AccessToken)

	return client, token, err
}

func (cfg *Config) Client(accessToken string) (mfclients.Client, error) {
	resp, err := http.Get(cfg.userInfoURL + url.QueryEscape(accessToken))
	if err != nil {
		return mfclients.Client{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return mfclients.Client{}, err
	}

	var user struct {
		ID    string `json:"sub"`
		Name  string `json:"preferred_username"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(data, &user); err != nil {
		return mfclients.Client{}, err
	}

	client := mfclients.Client{
		ID:   user.ID,
		Name: user.Name,
		Credentials: mfclients.Credentials{
			Identity: user.Email,
		},
		Metadata: map[string]interface{}{
			"provider": "kratos",
		},
		Status: mfclients.EnabledStatus,
	}

	return client, nil
}

func CallbackHandler(cfg *Config, svc users.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// state is prefixed with signin- or signup- to indicate which flow we should use
		var action, state string
		if r.FormValue("state") != "" && strings.Contains(r.FormValue("state"), "-") {
			action, state = strings.Split(r.FormValue("state"), "-")[0], strings.Split(r.FormValue("state"), "-")[1]
		}

		if state != cfg.state {
			http.Redirect(w, r, cfg.uiRedirectURL, http.StatusTemporaryRedirect)
			return
		}

		if code := r.FormValue("code"); code != "" {
			client, token, err := cfg.Profile(r.Context(), code)
			if err != nil {
				http.Redirect(w, r, cfg.uiRedirectURL, http.StatusTemporaryRedirect)
				return
			}

			jwt, err := svc.OAuthCallback(r.Context(), action, token, client)
			if err != nil {
				// We set the error cookie to be read by the frontend
				cookie := &http.Cookie{
					Name:    "error",
					Value:   err.Error(),
					Path:    "/",
					Expires: time.Now().Add(time.Second),
				}

				http.SetCookie(w, cookie)
			}

			if jwt.AccessToken != "" && jwt.RefreshToken != nil {
				accessTokenCookie := &http.Cookie{
					Name:     "token",
					Value:    jwt.AccessToken,
					Path:     "/",
					HttpOnly: true,
				}
				refresTokenCookie := &http.Cookie{
					Name:     "refresh_token",
					Value:    *jwt.RefreshToken,
					Path:     "/",
					HttpOnly: true,
				}

				http.SetCookie(w, accessTokenCookie)
				http.SetCookie(w, refresTokenCookie)
			}
		}

		http.Redirect(w, r, cfg.uiRedirectURL, http.StatusTemporaryRedirect)
	}
}
