// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package jwt

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/auth/kratos"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	errInvalidIssuer = errors.New("invalid token issuer value")
	// errJWTExpiryKey is used to check if the token is expired.
	errJWTExpiryKey = errors.New(`"exp" not satisfied`)
	// ErrExpiry indicates that the token is expired.
	ErrExpiry = errors.New("token is expired")
	// ErrSetClaim indicates an inability to set the claim.
	ErrSetClaim = errors.New("failed to set claim")
	// ErrSignJWT indicates an error in signing jwt token.
	ErrSignJWT = errors.New("failed to sign jwt token")
	// ErrParseToken indicates a failure to parse the token.
	ErrParseToken = errors.New("failed to parse token")
	// ErrValidateJWTToken indicates a failure to validate JWT token.
	ErrValidateJWTToken = errors.New("failed to validate jwt token")
	// ErrJSONHandle indicates an error in handling JSON.
	ErrJSONHandle = errors.New("failed to perform operation JSON")
)

const (
	issuerName              = "magistrala.auth"
	tokenType               = "type"
	userField               = "user"
	domainField             = "domain"
	kratosAccessTokenField  = "access_token"
	kratosRefreshTokenField = "refresh_token"
)

type tokenizer struct {
	secret []byte
	sdk    *kratos.SDK
}

var _ auth.Tokenizer = (*tokenizer)(nil)

// NewRepository instantiates an implementation of Token repository.
func New(secret []byte, sdk *kratos.SDK) auth.Tokenizer {
	return &tokenizer{
		secret: secret,
		sdk:    sdk,
	}
}

func (repo *tokenizer) Issue(key auth.Key) (string, error) {
	builder := jwt.NewBuilder()
	builder.
		Issuer(issuerName).
		IssuedAt(key.IssuedAt).
		Subject(key.Subject).
		Claim(tokenType, key.Type).
		Expiration(key.ExpiresAt)
	builder.Claim(userField, key.User)
	builder.Claim(domainField, key.Domain)
	builder.Claim(kratosAccessTokenField, key.Kratos.AccessToken)
	builder.Claim(kratosRefreshTokenField, key.Kratos.RefreshToken)
	if key.ID != "" {
		builder.JwtID(key.ID)
	}
	tkn, err := builder.Build()
	if err != nil {
		return "", errors.Wrap(svcerr.ErrAuthentication, err)
	}
	signedTkn, err := jwt.Sign(tkn, jwt.WithKey(jwa.HS512, repo.secret))
	if err != nil {
		return "", errors.Wrap(ErrSignJWT, err)
	}
	return string(signedTkn), nil
}

func (repo *tokenizer) Parse(token string) (auth.Key, error) {
	tkn, err := jwt.Parse(
		[]byte(token),
		jwt.WithValidate(true),
		jwt.WithKey(jwa.HS512, repo.secret),
	)
	if err != nil {
		if errors.Contains(err, errJWTExpiryKey) {
			return auth.Key{}, ErrExpiry
		}

		return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	validator := jwt.ValidatorFunc(func(_ context.Context, t jwt.Token) jwt.ValidationError {
		if t.Issuer() != issuerName {
			return jwt.NewValidationError(errInvalidIssuer)
		}
		return nil
	})
	if err := jwt.Validate(tkn, jwt.WithValidator(validator)); err != nil {
		return auth.Key{}, errors.Wrap(ErrValidateJWTToken, err)
	}

	jsn, err := json.Marshal(tkn.PrivateClaims())
	if err != nil {
		return auth.Key{}, errors.Wrap(ErrJSONHandle, err)
	}
	var key auth.Key
	if err := json.Unmarshal(jsn, &key); err != nil {
		return auth.Key{}, errors.Wrap(ErrJSONHandle, err)
	}

	tType, ok := tkn.Get(tokenType)
	if !ok {
		return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	ktype, err := strconv.ParseInt(fmt.Sprintf("%v", tType), 10, 64)
	if err != nil {
		return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}

	key.ID = tkn.JwtID()
	key.Type = auth.KeyType(ktype)
	key.Issuer = tkn.Issuer()
	key.Subject = tkn.Subject()
	key.IssuedAt = tkn.IssuedAt()
	key.ExpiresAt = tkn.Expiration()

	kratosAccessToken, ok := tkn.Get(kratosAccessTokenField)
	switch {
	case ok:
		switch repo.sdk.Validate(context.Background(), kratosAccessToken.(string)) {
		case nil:
			key.Kratos.AccessToken = kratosAccessToken.(string)
		default:
			kratosRefreshToken, ok := tkn.Get(kratosRefreshTokenField)
			if !ok {
				return auth.Key{}, errors.Wrap(errors.ErrAuthentication, err)
			}
			token, err := repo.sdk.Refresh(context.Background(), kratosRefreshToken.(string))
			if err != nil {
				return auth.Key{}, errors.Wrap(errors.ErrAuthentication, err)
			}
			key.Kratos.AccessToken = token.AccessToken
			key.Kratos.RefreshToken = token.RefreshToken

			return key, nil
		}
	case !ok:
		return auth.Key{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	kratosRefreshToken, ok := tkn.Get(kratosRefreshTokenField)
	switch {
	case ok:
		key.Kratos.RefreshToken = kratosRefreshToken.(string)
	case !ok:
		return auth.Key{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	return key, nil
}
