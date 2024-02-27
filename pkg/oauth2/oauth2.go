// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package oauth2

import (
	"context"
	"errors"

	mfclients "github.com/absmach/magistrala/pkg/clients"
	"golang.org/x/oauth2"
)

// State is the state of the OAuth2 flow.
type State uint8

const (
	// SignIn is the state for the sign-in flow.
	SignIn State = iota
	// SignUp is the state for the sign-up flow.
	SignUp
)

func (s State) String() string {
	switch s {
	case SignIn:
		return "signin"
	case SignUp:
		return "signup"
	default:
		return "unknown"
	}
}

// ToState converts string value to a valid OAuth2 state.
func ToState(state string) (State, error) {
	switch state {
	case "signin":
		return SignIn, nil
	case "signup":
		return SignUp, nil
	}

	return State(0), errors.New("invalid state")
}

// Config is the configuration for the OAuth2 provider.
type Config struct {
	ClientID     string `env:"CLIENT_ID"       envDefault:""`
	ClientSecret string `env:"CLIENT_SECRET"   envDefault:""`
	State        string `env:"STATE"           envDefault:""`
	RedirectURL  string `env:"REDIRECT_URL"    envDefault:""`
}

// Provider is an interface that provides the OAuth2 flow for a specific provider
// (e.g. Google, GitHub, etc.)
//
//go:generate mockery --name Provider --output=./mocks --filename oauth2.go --quiet --note "Copyright (c) Abstract Machines"
type Provider interface {
	// Name returns the name of the OAuth2 provider
	Name() string
	// State returns the state for the OAuth2 flow
	State() string
	// RedirectURL returns the URL to redirect the user to after the OAuth2 flow
	RedirectURL() string
	// ErrorURL returns the URL to redirect the user to if an error occurs during the OAuth2 flow
	ErrorURL() string
	// IsEnabled returns true if the provider is enabled
	IsEnabled() bool
	// Profile returns the user's profile and token from the OAuth2 provider
	Profile(ctx context.Context, code string) (mfclients.Client, oauth2.Token, error)
	// Validate checks if the access token is valid.
	Validate(ctx context.Context, token string) error
	// Refresh refreshes the token.
	Refresh(ctx context.Context, token string) (oauth2.Token, error)
}
