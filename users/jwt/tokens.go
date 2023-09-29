// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package jwt

import (
	"context"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/mainflux/mainflux/pkg/errors"
	"google.golang.org/api/idtoken"
)

const (
	mainfluxIssuer = "mainflux.users"
	googleIssuer   = "https://accounts.google.com"
)

var _ Repository = (*tokenRepo)(nil)

type tokenRepo struct {
	secret          []byte
	googleClientID  string
	accessDuration  time.Duration
	refreshDuration time.Duration
}

// NewRepository instantiates an implementation of Token repository.
func NewRepository(secret []byte, aduration, rduration time.Duration, gClientID string) Repository {
	return &tokenRepo{
		secret:          secret,
		googleClientID:  gClientID,
		accessDuration:  aduration,
		refreshDuration: rduration,
	}
}

func (repo tokenRepo) Issue(ctx context.Context, claim Claims) (Token, error) {
	aexpiry := time.Now().Add(repo.accessDuration)
	accessToken, err := jwt.NewBuilder().
		Issuer(mainfluxIssuer).
		IssuedAt(time.Now()).
		Subject(claim.ClientID).
		Claim("type", AccessToken).
		Expiration(aexpiry).
		Build()
	if err != nil {
		return Token{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	signedAccessToken, err := jwt.Sign(accessToken, jwt.WithKey(jwa.HS512, repo.secret))
	if err != nil {
		return Token{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	refreshToken, err := jwt.NewBuilder().
		Issuer(mainfluxIssuer).
		IssuedAt(time.Now()).
		Subject(claim.ClientID).
		Claim("type", RefreshToken).
		Expiration(time.Now().Add(repo.refreshDuration)).
		Build()
	if err != nil {
		return Token{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	signedRefreshToken, err := jwt.Sign(refreshToken, jwt.WithKey(jwa.HS512, repo.secret))
	if err != nil {
		return Token{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	return Token{
		AccessToken:  string(signedAccessToken),
		RefreshToken: string(signedRefreshToken),
		AccessType:   "Bearer",
	}, nil
}

func (repo tokenRepo) Parse(ctx context.Context, accessToken string) (Claims, error) {
	// Decode token to get issuer
	token, err := jwt.Parse(
		[]byte(accessToken),
		jwt.WithValidate(true),
		jwt.WithVerify(false),
	)
	if err != nil {
		return Claims{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	switch token.Issuer() {
	case mainfluxIssuer:
		token, err := jwt.Parse(
			[]byte(accessToken),
			jwt.WithValidate(true),
			jwt.WithKey(jwa.HS512, repo.secret),
		)
		if err != nil {
			return Claims{}, errors.Wrap(errors.ErrAuthentication, err)
		}
		tType, ok := token.Get("type")
		if !ok {
			return Claims{}, errors.Wrap(errors.ErrAuthentication, err)
		}
		claim := Claims{
			ClientID: token.Subject(),
			Type:     tType.(string),
		}
		return claim, nil
	case googleIssuer:
		tokenValidator, err := idtoken.NewValidator(ctx)
		if err != nil {
			return Claims{}, errors.Wrap(errors.ErrAuthentication, err)
		}

		payload, err := tokenValidator.Validate(ctx, accessToken, repo.googleClientID)
		if err != nil {
			return Claims{}, errors.Wrap(errors.ErrAuthentication, err)
		}

		return Claims{
			ClientID: payload.Subject,
			Type:     AccessToken,
		}, nil
	default:
		return Claims{}, errors.Wrap(errors.ErrAuthentication, err)
	}
}
