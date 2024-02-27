// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package oauth

import (
	"context"
	"net/http"
	"strings"

	mfclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/users"
	"golang.org/x/oauth2"
)

// Config is the configuration for the OAuth2 provider.
type Config struct {
	ClientID     string `env:"CLIENT_ID"       envDefault:""`
	ClientSecret string `env:"CLIENT_SECRET"   envDefault:""`
	State        string `env:"STATE"           envDefault:""`
	RedirectURL  string `env:"REDIRECT_URL"    envDefault:""`
}

// Provider is an interface that provides the OAuth2 flow for a specific provider
// (e.g. Google, GitHub, etc.)
type Provider interface {
	// Name returns the name of the OAuth2 provider
	Name() string
	// State returns the state for the OAuth2 flow
	State() string
	// RedirectURL returns the URL to redirect the user to after the OAuth2 flow
	RedirectURL() string
	// ErrorURL returns the URL to redirect the user to if an error occurs during the OAuth2 flow
	ErrorURL() string
	// Profile returns the user's profile and token from the OAuth2 provider
	Profile(ctx context.Context, code string) (mfclients.Client, oauth2.Token, error)
	// Validate checks if the access token is valid.
	Validate(ctx context.Context, token string) error
	// Refresh refreshes the token.
	Refresh(ctx context.Context, token string) (oauth2.Token, error)
}

// CallbackHandler is a http.HandlerFunc that handles OAuth2 callbacks.
func CallbackHandler(oauth Provider, svc users.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// state is prefixed with signin- or signup- to indicate which flow we should use
		var action, state string
		if r.FormValue("state") != "" && strings.Contains(r.FormValue("state"), "-") {
			action, state = strings.Split(r.FormValue("state"), "-")[0], strings.Split(r.FormValue("state"), "-")[1]
		}

		if state != oauth.State() {
			http.Redirect(w, r, oauth.ErrorURL()+"?error=invalid%20state", http.StatusSeeOther)
			return
		}

		if code := r.FormValue("code"); code != "" {
			client, token, err := oauth.Profile(r.Context(), code)
			if err != nil {
				http.Redirect(w, r, oauth.ErrorURL()+"?error="+err.Error(), http.StatusSeeOther)
				return
			}

			jwt, err := svc.OAuthCallback(r.Context(), oauth.Name(), action, token, client)
			if err != nil {
				http.Redirect(w, r, oauth.ErrorURL()+"?error="+err.Error(), http.StatusSeeOther)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "access_token",
				Value:    jwt.AccessToken,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			})
			http.SetCookie(w, &http.Cookie{
				Name:     "refresh_token",
				Value:    *jwt.RefreshToken,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			})

			http.Redirect(w, r, oauth.RedirectURL(), http.StatusFound)
			return
		}

		http.Redirect(w, r, oauth.ErrorURL()+"?error=empty%20code", http.StatusSeeOther)
	}
}
