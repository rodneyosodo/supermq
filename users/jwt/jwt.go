// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package jwt

import (
	"context"
)

const (
	AccessToken  string = "access"
	RefreshToken string = "refresh"
	RestPassword string = "reset"
	Invitation   string = "invitation"
)

// Token is used for authentication purposes.
// It contains AccessToken, RefreshToken, Type and AccessExpiry.
type Token struct {
	AccessToken  string // AccessToken contains the security credentials for a login session and identifies the client.
	RefreshToken string // RefreshToken is a credential artifact that OAuth can use to get a new access token without client interaction.
	AccessType   string // AccessType is the specific type of access token issued.
}

// Claims are the Client's internal JWT Claims.
type Claims struct {
	ClientID string // ClientID is the client unique identifier.
	Type     string // Type denotes the type of claim.
}

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// IssueToken issues a new access and refresh token.
	IssueToken(ctx context.Context, identity, secret string) (Token, error)

	// RefreshToken refreshes expired access tokens.
	// After an access token expires, the refresh token is used to get
	// a new pair of access and refresh tokens.
	RefreshToken(ctx context.Context, accessToken string) (Token, error)
}

// Repository specifies an account persistence API.
type Repository interface {
	// Issue issues a new access and refresh token.
	Issue(ctx context.Context, claim Claims) (Token, error)

	// Parse checks the validity of a token.
	Parse(ctx context.Context, token string) (Claims, error)
}
