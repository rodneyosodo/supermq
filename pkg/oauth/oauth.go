// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package oauth

import (
	"context"
	"net/http"
	"strings"
	"time"

	mfclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/errors"
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
			encodeError(w, errors.New("invalid state"))
			http.Redirect(w, r, oauth.RedirectURL(), http.StatusSeeOther)
			return
		}

		if code := r.FormValue("code"); code != "" {
			client, token, err := oauth.Profile(r.Context(), code)
			if err != nil {
				encodeError(w, err)
				http.Redirect(w, r, oauth.RedirectURL(), http.StatusSeeOther)
				return
			}

			jwt, err := svc.OAuthCallback(r.Context(), oauth.Name(), action, token, client)
			if err != nil {
				encodeError(w, err)
				http.Redirect(w, r, oauth.RedirectURL(), http.StatusSeeOther)
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

		encodeError(w, errors.New("empty code"))
		http.Redirect(w, r, oauth.RedirectURL(), http.StatusSeeOther)
	}
}

func encodeError(w http.ResponseWriter, err error) {
	http.SetCookie(w, &http.Cookie{
		Name:    "magistrala_error",
		Value:   err.Error(),
		Path:    "/",
		Expires: time.Now().Add(time.Second * 10),
	})
}
