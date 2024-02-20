// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Package kratos contains kratos SDK.
package kratos

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/absmach/magistrala/pkg/errors"
	ory "github.com/ory/client-go"
	"golang.org/x/oauth2"
)

const TokenEndpoint = "/oauth2/token"

type SDK struct {
	client       *ory.APIClient
	url          string
	clientID     string
	clientSecret string
}

func NewSDK(url, apiKey, clientID, clientSecret string) *SDK {
	conf := ory.NewConfiguration()
	conf.Servers = []ory.ServerConfiguration{{URL: url}}
	conf.AddDefaultHeader("Authorization", "Bearer "+apiKey)
	client := ory.NewAPIClient(conf)

	return &SDK{
		url:          url,
		client:       client,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// Validate checks if the token is valid.
func (sdk *SDK) Validate(ctx context.Context, token string) error {
	introspectedToken, resp, err := sdk.client.OAuth2API.IntrospectOAuth2Token(ctx).Token(token).Execute()
	if err != nil {
		return fmt.Errorf("failed to identify user: %w", decodeError(resp))
	}
	if !introspectedToken.Active {
		return errors.ErrAuthentication
	}

	return nil
}

// Refresh refreshes the token.
func (sdk *SDK) Refresh(ctx context.Context, token string) (oauth2.Token, error) {
	payload := strings.NewReader("grant_type=refresh_token&refresh_token=" + token + "&scope=email%20profile%20offline_access")

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, sdk.url+TokenEndpoint, payload)
	if err != nil {
		return oauth2.Token{}, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+basicAuth(sdk.clientID, sdk.clientSecret))

	res, err := client.Do(req)
	if err != nil {
		return oauth2.Token{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return oauth2.Token{}, errors.ErrAuthentication
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return oauth2.Token{}, err
	}
	var tokenData oauth2.Token
	if err := json.Unmarshal(body, &tokenData); err != nil {
		return oauth2.Token{}, err
	}

	return tokenData, nil
}

func basicAuth(id, secret string) string {
	auth := id + ":" + secret
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func decodeError(response *http.Response) error {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	var content struct {
		Error ory.GenericError `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &content); err != nil {
		return fmt.Errorf("error unmarshalling response body: %w", err)
	}

	return fmt.Errorf("error: %s, reason: %s", content.Error.Message, *content.Error.Reason)
}
