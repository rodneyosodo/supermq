package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/users/clients"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
)

var (
	scopes = []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	}

	errFailedToSignUp  = errors.New("failed to sign up")
	errUserNotSignedUp = errors.New("user not signed up")

	cookieErrDuration = 10 * time.Second
)

type Config struct {
	*oauth2.Config
	State string
}

type user struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

func NewConfig(clientId, clientSecret, redirectURL string) Config {
	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	return Config{Config: cfg}
}

func (conf *Config) Profile(ctx context.Context, code string) (mfclients.Client, *oauth2.Token, error) {
	token, err := conf.Config.Exchange(ctx, code)
	if err != nil {
		return mfclients.Client{}, &oauth2.Token{}, err
	}

	resp, err := http.Get(userInfoURL + url.QueryEscape(token.AccessToken))
	if err != nil {
		return mfclients.Client{}, token, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return mfclients.Client{}, token, err
	}

	var user user
	if err := json.Unmarshal(data, &user); err != nil {
		return mfclients.Client{}, token, err
	}

	var client = mfclients.Client{
		ID:   user.ID,
		Name: user.Name,
		Credentials: mfclients.Credentials{
			Identity: user.Email,
		},
		Metadata: map[string]interface{}{
			"profile_picture": user.Picture,
			"provider":        "google",
		},
		Status: mfclients.EnabledStatus,
	}

	return client, token, nil
}

func CallbackHandler(conf *Config, svc clients.Service, redirectURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// state is prefixed with signin- or signup- to indicate which flow we should use
		var action, state string
		if r.FormValue("state") != "" && strings.Contains(r.FormValue("state"), "-") {
			action, state = strings.Split(r.FormValue("state"), "-")[0], strings.Split(r.FormValue("state"), "-")[1]
		}

		if state != conf.State {
			http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
			return
		}

		if code := r.FormValue("code"); code != "" {
			client, token, err := conf.Profile(r.Context(), code)
			if err != nil {
				http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
				return
			}

			var idToken = token.Extra("id_token").(string)

			var rclient mfclients.Client
			switch action {
			case "signin":
				rclient, err = svc.ViewProfile(r.Context(), idToken)
				if err != nil {
					err = errUserNotSignedUp
				}
			case "signup":
				rclient, err = svc.RegisterClient(r.Context(), "", client)
				if err != nil {
					err = errFailedToSignUp
				}
			default:
				err = nil
			}

			if err != nil {
				// We set the error cookie to be read by the frontend
				cookie := &http.Cookie{
					Name:    "error",
					Value:   err.Error(),
					Path:    "/",
					Expires: time.Now().Add(cookieErrDuration),
				}

				http.SetCookie(w, cookie)
			}

			if client.ID == rclient.ID {
				cookie := &http.Cookie{
					Name:     "token",
					Value:    idToken,
					Path:     "/",
					HttpOnly: true,
				}

				http.SetCookie(w, cookie)
			}
		}

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}
