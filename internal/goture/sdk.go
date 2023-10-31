// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	mfclients "github.com/mainflux/mainflux/pkg/clients"
)

const (
	// DefaultBaseURL is the default API address of the SDK.
	DefaultBaseURL = "http://gotrue:9999"
)

type User struct {
	ID               string                   `json:"id"`
	Aud              string                   `json:"aud"`
	Role             string                   `json:"role"`
	Email            string                   `json:"email"`
	EmailConfirmedAt string                   `json:"email_confirmed_at"`
	Phone            string                   `json:"phone"`
	ConfirmedAt      string                   `json:"confirmed_at"`
	LastSignInAt     string                   `json:"last_sign_in_at"`
	AppMetadata      map[string]interface{}   `json:"app_metadata"`
	UserMetadata     map[string]interface{}   `json:"user_metadata"`
	Identities       []map[string]interface{} `json:"identities"`
	CreatedAt        string                   `json:"created_at"`
	UpdatedAt        string                   `json:"updated_at"`
}

func IdentifyUser(token string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", DefaultBaseURL+"/user", nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return "", err
	}

	return user.ID, nil
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

func IssueToken(email, password string) (TokenResponse, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return TokenResponse{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", DefaultBaseURL+"/token?grant_type=password", bytes.NewReader(reqBody))
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

func RefreshToken(refreshToken string) (TokenResponse, error) {
	payload := map[string]string{
		"refresh_token": refreshToken,
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return TokenResponse{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", DefaultBaseURL+"/token?grant_type=refresh_token", bytes.NewReader(reqBody))
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

func CreateUser(cli mfclients.Client) (mfclients.Client, error) {
	user := struct {
		Email    string                 `json:"email"`
		Password string                 `json:"password"`
		Data     map[string]interface{} `json:"data"`
	}{
		Email:    cli.Credentials.Identity,
		Password: cli.Credentials.Secret,
		Data:     cli.Metadata,
	}

	payload, err := json.Marshal(user)
	if err != nil {
		return mfclients.Client{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", DefaultBaseURL+"/signup", bytes.NewReader(payload))
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

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return mfclients.Client{}, err
	}

	cli = mfclients.Client{
		ID: token.User.ID,
		Credentials: mfclients.Credentials{
			Identity: token.User.Email,
		},
		Metadata: token.User.UserMetadata,
		Status:   mfclients.EnabledStatus,
	}

	cli.CreatedAt, err = time.Parse(time.RFC3339Nano, token.User.CreatedAt)
	if err != nil {
		return cli, err
	}
	cli.UpdatedAt, err = time.Parse(time.RFC3339Nano, token.User.UpdatedAt)
	if err != nil {
		return cli, err
	}

	return cli, nil
}

type ListUsersResponse struct {
	Users []User `json:"users"`
	Aud   string `json:"aud"`
}

func ListUsers(token string, offset, limit uint64) (mfclients.ClientsPage, error) {
	offset = offset + 1
	client := &http.Client{}
	req, err := http.NewRequest("GET", DefaultBaseURL+"/admin/users?page="+fmt.Sprint(offset)+"&per_page="+fmt.Sprint(limit), nil)
	if err != nil {
		return mfclients.ClientsPage{}, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		return mfclients.ClientsPage{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return mfclients.ClientsPage{}, err
	}

	var users ListUsersResponse
	if err := json.Unmarshal(body, &users); err != nil {
		return mfclients.ClientsPage{}, err
	}

	return toClientsPage(users), nil
}

func toClientsPage(users ListUsersResponse) mfclients.ClientsPage {
	clients := make([]mfclients.Client, len(users.Users))
	for i, user := range users.Users {
		createdAt, err := time.Parse(time.RFC3339Nano, user.CreatedAt)
		if err != nil {
			return mfclients.ClientsPage{}
		}
		updatedAt, err := time.Parse(time.RFC3339Nano, user.UpdatedAt)
		if err != nil {
			return mfclients.ClientsPage{}
		}
		clients[i] = mfclients.Client{
			ID: user.ID,
			Credentials: mfclients.Credentials{
				Identity: user.Email,
			},
			Metadata:  user.UserMetadata,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			Status:    mfclients.EnabledStatus,
		}
	}
	return mfclients.ClientsPage{
		Page: mfclients.Page{
			Total: uint64(len(users.Users)),
		},
		Clients: clients,
	}
}

func GetUser(token, id string) (mfclients.Client, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", DefaultBaseURL+"/admin/users/"+id, nil)
	if err != nil {
		return mfclients.Client{}, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		return mfclients.Client{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return mfclients.Client{}, err
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return mfclients.Client{}, err
	}

	createdAt, err := time.Parse(time.RFC3339Nano, user.CreatedAt)
	if err != nil {
		return mfclients.Client{}, err
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, user.UpdatedAt)
	if err != nil {
		return mfclients.Client{}, err
	}

	return mfclients.Client{
		ID: user.ID,
		Credentials: mfclients.Credentials{
			Identity: user.Email,
		},
		Metadata:  user.UserMetadata,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Status:    mfclients.EnabledStatus,
	}, nil
}

func UpdateUser(token string, cli mfclients.Client) (mfclients.Client, error) {
	user := User{
		ID:           cli.ID,
		Email:        cli.Credentials.Identity,
		UserMetadata: cli.Metadata,
	}
	payload, err := json.Marshal(user)
	if err != nil {
		return mfclients.Client{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("PUT", DefaultBaseURL+"/admin/users/"+user.ID, bytes.NewReader(payload))
	if err != nil {
		return mfclients.Client{}, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		return mfclients.Client{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return mfclients.Client{}, err
	}

	user = User{}
	if err := json.Unmarshal(body, &user); err != nil {
		return mfclients.Client{}, err
	}

	createdAt, err := time.Parse(time.RFC3339Nano, user.CreatedAt)
	if err != nil {
		return mfclients.Client{}, err
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, user.UpdatedAt)
	if err != nil {
		return mfclients.Client{}, err
	}
	return mfclients.Client{
		ID: user.ID,
		Credentials: mfclients.Credentials{
			Identity: user.Email,
		},
		Metadata:  user.UserMetadata,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Status:    mfclients.EnabledStatus,
	}, nil
}
