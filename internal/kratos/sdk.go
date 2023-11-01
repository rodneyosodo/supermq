// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/slices"
)

const (
	// DefaultBaseURL is the default API address of the SDK.
	DefaultBaseURL = "http://kratos:4434"
)

type user struct {
	ID          string `json:"id"`
	Credentials struct {
		Password credentialsResponse `json:"password"`
		Webauthn credentialsResponse `json:"webauthn"`
	} `json:"credentials"`
	SchemaID    string                 `json:"schema_id"`
	SchemaURL   string                 `json:"schema_url"`
	State       string                 `json:"state"`
	StateChange string                 `json:"state_changed_at"`
	Traits      traits                 `json:"traits"`
	Verifiable  []verifiableAddresses  `json:"verifiable_addresses"`
	Recovery    []recoveryAddresses    `json:"recovery_addresses"`
	Metadata    map[string]interface{} `json:"metadata_public"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

type session struct {
	ID                    string                   `json:"id"`
	Active                bool                     `json:"active"`
	ExpiresAt             string                   `json:"expires_at"`
	AuthenticatedAt       string                   `json:"authenticated_at"`
	AAL                   string                   `json:"authenticator_assurance_level"`
	AuthenticationMethods []map[string]interface{} `json:"authentication_methods"`
	IssuedAt              string                   `json:"issued_at"`
	Identity              user                     `json:"identity"`
	Devices               []map[string]interface{} `json:"devices"`
}

func IdentifyUser(token string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", strings.Replace(DefaultBaseURL, "4434", "4433", 1)+"/sessions/whoami", nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Session-Token", token)

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var user session
	if err := json.Unmarshal(body, &user); err != nil {
		return "", err
	}

	return user.Identity.ID, nil
}

type LoginFlow struct {
	ID                   string                 `json:"id"`
	OAuth2LoginChallenge interface{}            `json:"oauth2_login_challenge"`
	Type                 string                 `json:"type"`
	ExpiresAt            string                 `json:"expires_at"`
	IssuedAt             string                 `json:"issued_at"`
	RequestURL           string                 `json:"request_url"`
	UI                   map[string]interface{} `json:"ui"`
	CreatedAt            string                 `json:"created_at"`
	UpdatedAt            string                 `json:"updated_at"`
	Refresh              bool                   `json:"refresh"`
	RequestedAAL         string                 `json:"requested_aal"`
}

func createLoginFlow() (LoginFlow, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", DefaultBaseURL+"/self-service/login/api?refresh=true&aal=aal1&return_session_token_exchange_code=true", nil)
	if err != nil {
		fmt.Printf("failed to create request: %s\n", err)
		return LoginFlow{}, err
	}

	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("failed to send request: %s\n", err)
		return LoginFlow{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("failed to read response body: %s\n", err)
		return LoginFlow{}, err
	}

	var loginFlow LoginFlow
	if err := json.Unmarshal(body, &loginFlow); err != nil {
		fmt.Printf("failed to unmarshal response body: %s\n", err)
		return LoginFlow{}, err
	}

	return loginFlow, nil
}

type TokenResponse struct {
	SessionToken string  `json:"session_token"`
	Session      session `json:"session"`
}

func IssueToken(email, password string) (TokenResponse, error) {
	loginFlow, err := createLoginFlow()
	if err != nil {
		return TokenResponse{}, err
	}

	payload := map[string]string{
		"identifier":          email,
		"password":            password,
		"method":              "password",
		"password_identifier": email,
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return TokenResponse{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", DefaultBaseURL+"/self-service/login?flow="+loginFlow.ID, bytes.NewReader(reqBody))
	if err != nil {
		return TokenResponse{}, err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return TokenResponse{}, err
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return TokenResponse{}, err
	}

	return token, nil
}

type traits struct {
	Email      string `json:"email"`
	Username   string `json:"username"`
	Enterprise bool   `json:"enterprise"`
	Newsletter bool   `json:"newsletter"`
}

type credentials struct {
	Password struct {
		Config struct {
			HashedPassword string `json:"hashed_password"`
			Password       string `json:"password"`
		} `json:"config"`
	} `json:"password"`
}

type recoveryAddresses struct {
	Value string `json:"value"`
	Via   string `json:"via"`
}

type verifiableAddresses struct {
	Verified bool   `json:"verified"`
	Value    string `json:"value"`
	Status   string `json:"status"`
	Via      string `json:"via"`
}

type createUserRequest struct {
	SchemaID            string                `json:"schema_id"`
	Traits              traits                `json:"traits"`
	Credentials         credentials           `json:"credentials"`
	RecoveryAddresses   []recoveryAddresses   `json:"recovery_addresses"`
	State               string                `json:"state"`
	VerifiableAddresses []verifiableAddresses `json:"verifiable_addresses"`
}

type credentialsResponse struct {
	Type        string   `json:"type"`
	Identifiers []string `json:"identifiers"`
	Version     int      `json:"version"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

func CreateUser(cli mfclients.Client) (mfclients.Client, error) {
	cureq := createUserRequest{
		SchemaID: "default",
		Traits: traits{
			Email:      cli.Credentials.Identity,
			Username:   cli.Name,
			Enterprise: slices.Contains(cli.Tags, "enterprise"),
			Newsletter: slices.Contains(cli.Tags, "newsletter"),
		},
		Credentials: credentials{
			Password: struct {
				Config struct {
					HashedPassword string `json:"hashed_password"`
					Password       string `json:"password"`
				} `json:"config"`
			}{
				Config: struct {
					HashedPassword string `json:"hashed_password"`
					Password       string `json:"password"`
				}{
					HashedPassword: hashPassword(cli.Credentials.Secret),
					Password:       cli.Credentials.Secret,
				},
			},
		},
		RecoveryAddresses: []recoveryAddresses{
			{
				Value: cli.Credentials.Identity,
				Via:   "email",
			},
		},
		State: "active",
		VerifiableAddresses: []verifiableAddresses{
			{
				Verified: cli.Status == mfclients.EnabledStatus,
				Value:    cli.Credentials.Identity,
				Status:   "completed",
				Via:      "email",
			},
		},
	}

	payload, err := json.Marshal(cureq)
	if err != nil {
		return mfclients.Client{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", DefaultBaseURL+"/admin/identities", bytes.NewReader(payload))
	if err != nil {
		return mfclients.Client{}, err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return mfclients.Client{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return mfclients.Client{}, err
	}

	var curesp user
	if err := json.Unmarshal(body, &curesp); err != nil {
		return mfclients.Client{}, err
	}

	return toClients(curesp), nil
}

func ListUsers(token string, offset, limit uint64) (mfclients.ClientsPage, error) {
	offset = offset + 1
	client := &http.Client{}
	req, err := http.NewRequest("GET", DefaultBaseURL+"/admin/identities?page="+fmt.Sprint(offset)+"&per_page="+fmt.Sprint(limit), nil)
	if err != nil {
		return mfclients.ClientsPage{}, err
	}

	req.Header.Add("Authorization", token)

	res, err := client.Do(req)
	if err != nil {
		return mfclients.ClientsPage{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return mfclients.ClientsPage{}, err
	}

	var users []user
	if err := json.Unmarshal(body, &users); err != nil {
		return mfclients.ClientsPage{}, err
	}

	return toClientsPage(users), nil
}

func toClientsPage(users []user) mfclients.ClientsPage {
	clients := make([]mfclients.Client, len(users))
	for i, user := range users {
		clients[i] = toClients(user)
	}

	return mfclients.ClientsPage{
		Page: mfclients.Page{
			Total: uint64(len(users)),
		},
		Clients: clients,
	}
}

func toClients(user user) mfclients.Client {
	createdAt, err := time.Parse(time.RFC3339Nano, user.CreatedAt)
	if err != nil {
		return mfclients.Client{}
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, user.UpdatedAt)
	if err != nil {
		return mfclients.Client{}
	}
	var tags []string
	if user.Traits.Enterprise {
		tags = append(tags, "enterprise")
	}
	if user.Traits.Newsletter {
		tags = append(tags, "newsletter")
	}
	return mfclients.Client{
		ID:   user.ID,
		Name: user.Traits.Username,
		Tags: tags,
		Credentials: mfclients.Credentials{
			Identity: user.Traits.Email,
		},
		Metadata:  user.Metadata,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Status:    mfclients.EnabledStatus,
	}
}

func GetUser(token, id string) (mfclients.Client, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", DefaultBaseURL+"/admin/identities/"+id, nil)
	if err != nil {
		return mfclients.Client{}, err
	}

	req.Header.Add("Authorization", token)

	res, err := client.Do(req)
	if err != nil {
		return mfclients.Client{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return mfclients.Client{}, err
	}

	var user user
	if err := json.Unmarshal(body, &user); err != nil {
		return mfclients.Client{}, err
	}

	return toClients(user), nil
}

func UpdateUser(token string, cli mfclients.Client) (mfclients.Client, error) {
	cuReq := user{
		SchemaID: "default",
		Traits: traits{
			Email:      cli.Credentials.Identity,
			Username:   cli.Name,
			Enterprise: slices.Contains(cli.Tags, "enterprise"),
			Newsletter: slices.Contains(cli.Tags, "newsletter"),
		},
		Metadata: cli.Metadata,
	}
	payload, err := json.Marshal(cuReq)
	if err != nil {
		return mfclients.Client{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("PUT", DefaultBaseURL+"/admin/identities/"+cli.ID, bytes.NewReader(payload))
	if err != nil {
		return mfclients.Client{}, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)

	res, err := client.Do(req)
	if err != nil {
		return mfclients.Client{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return mfclients.Client{}, err
	}

	var user user
	if err := json.Unmarshal(body, &user); err != nil {
		return mfclients.Client{}, err
	}

	return toClients(user), nil
}

func hashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}
