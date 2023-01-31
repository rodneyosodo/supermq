// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux/auth/keys"
	httpapi "github.com/mainflux/mainflux/auth/keys/api"
	"github.com/mainflux/mainflux/auth/keys/jwt"
	"github.com/mainflux/mainflux/auth/keys/mocks"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	secret        = "secret"
	contentType   = "application/json"
	id            = "123e4567-e89b-12d3-a456-000000000001"
	email         = "user@example.com"
	loginDuration = 30 * time.Minute
)

type issueRequest struct {
	Duration time.Duration `json:"duration,omitempty"`
	Type     uint32        `json:"type,omitempty"`
}

type testRequest struct {
	client      *http.Client
	method      string
	url         string
	contentType string
	token       string
	body        io.Reader
}

func (tr testRequest) make() (*http.Response, error) {
	req, err := http.NewRequest(tr.method, tr.url, tr.body)
	if err != nil {
		return nil, err
	}
	if tr.token != "" {
		req.Header.Set("Authorization", apiutil.BearerPrefix+tr.token)
	}
	if tr.contentType != "" {
		req.Header.Set("Content-Type", tr.contentType)
	}

	req.Header.Set("Referer", "http://localhost")
	return tr.client.Do(req)
}

func newService() keys.Service {
	repo := mocks.NewKeyRepository()
	idProvider := uuid.NewMock()
	t := jwt.New(secret)

	return keys.NewService(repo, idProvider, t, loginDuration)
}

func newServer(svc keys.Service) *httptest.Server {
	logger := logger.NewMock()
	mux := bone.New()
	httpapi.MakeHandler(svc, mux, logger)
	return httptest.NewServer(mux)
}

func toJSON(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

func TestIssue(t *testing.T) {
	svc := newService()
	_, loginSecret, err := svc.Issue(context.Background(), "", keys.Key{Type: keys.LoginKey, IssuedAt: time.Now(), IssuerID: id, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()

	lk := issueRequest{Type: keys.LoginKey}
	ak := issueRequest{Type: keys.APIKey, Duration: time.Hour}
	rk := issueRequest{Type: keys.RecoveryKey}

	cases := []struct {
		desc   string
		req    string
		ct     string
		token  string
		status int
	}{
		{
			desc:   "issue login key with empty token",
			req:    toJSON(lk),
			ct:     contentType,
			token:  "",
			status: http.StatusUnauthorized,
		},
		{
			desc:   "issue API key",
			req:    toJSON(ak),
			ct:     contentType,
			token:  loginSecret,
			status: http.StatusCreated,
		},
		{
			desc:   "issue recovery key",
			req:    toJSON(rk),
			ct:     contentType,
			token:  loginSecret,
			status: http.StatusCreated,
		},
		{
			desc:   "issue login key wrong content type",
			req:    toJSON(lk),
			ct:     "",
			token:  loginSecret,
			status: http.StatusUnsupportedMediaType,
		},
		{
			desc:   "issue recovery key wrong content type",
			req:    toJSON(rk),
			ct:     "",
			token:  loginSecret,
			status: http.StatusUnsupportedMediaType,
		},
		{
			desc:   "issue key with an invalid token",
			req:    toJSON(ak),
			ct:     contentType,
			token:  "wrong",
			status: http.StatusUnauthorized,
		},
		{
			desc:   "issue recovery key with empty token",
			req:    toJSON(rk),
			ct:     contentType,
			token:  "",
			status: http.StatusUnauthorized,
		},
		{
			desc:   "issue key with invalid request",
			req:    "{",
			ct:     contentType,
			token:  loginSecret,
			status: http.StatusBadRequest,
		},
		{
			desc:   "issue key with invalid JSON",
			req:    "{invalid}",
			ct:     contentType,
			token:  loginSecret,
			status: http.StatusBadRequest,
		},
		{
			desc:   "issue key with invalid JSON content",
			req:    `{"Type":{"key":"value"}}`,
			ct:     contentType,
			token:  loginSecret,
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      client,
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/keys", ts.URL),
			contentType: tc.ct,
			token:       tc.token,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestRetrieve(t *testing.T) {
	svc := newService()
	_, loginSecret, err := svc.Issue(context.Background(), "", keys.Key{Type: keys.LoginKey, IssuedAt: time.Now(), IssuerID: id, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))
	key := keys.Key{Type: keys.APIKey, IssuedAt: time.Now(), IssuerID: id, Subject: email}

	k, _, err := svc.Issue(context.Background(), loginSecret, key)
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()

	cases := []struct {
		desc   string
		id     string
		token  string
		status int
	}{
		{
			desc:   "retrieve an existing key",
			id:     k.ID,
			token:  loginSecret,
			status: http.StatusOK,
		},
		{
			desc:   "retrieve a non-existing key",
			id:     "non-existing",
			token:  loginSecret,
			status: http.StatusNotFound,
		},
		{
			desc:   "retrieve a key with an invalid token",
			id:     k.ID,
			token:  "wrong",
			status: http.StatusUnauthorized,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client: client,
			method: http.MethodGet,
			url:    fmt.Sprintf("%s/keys/%s", ts.URL, tc.id),
			token:  tc.token,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestRevoke(t *testing.T) {
	svc := newService()
	_, loginSecret, err := svc.Issue(context.Background(), "", keys.Key{Type: keys.LoginKey, IssuedAt: time.Now(), IssuerID: id, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))
	key := keys.Key{Type: keys.APIKey, IssuedAt: time.Now(), IssuerID: id, Subject: email}

	k, _, err := svc.Issue(context.Background(), loginSecret, key)
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()

	cases := []struct {
		desc   string
		id     string
		token  string
		status int
	}{
		{
			desc:   "revoke an existing key",
			id:     k.ID,
			token:  loginSecret,
			status: http.StatusNoContent,
		},
		{
			desc:   "revoke a non-existing key",
			id:     "non-existing",
			token:  loginSecret,
			status: http.StatusNoContent,
		},
		{
			desc:   "revoke key with invalid token",
			id:     k.ID,
			token:  "wrong",
			status: http.StatusUnauthorized},
	}

	for _, tc := range cases {
		req := testRequest{
			client: client,
			method: http.MethodDelete,
			url:    fmt.Sprintf("%s/keys/%s", ts.URL, tc.id),
			token:  tc.token,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}