// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/policies"
	"golang.org/x/sync/errgroup"
)

type callback struct {
	httpClient *http.Client
	urls       []string
	method     string
}

// CallBack send auth request to an external service.
//
//go:generate mockery --name CallBack --output=./mocks --filename callback.go --quiet --note "Copyright (c) Abstract Machines"
type CallBack interface {
	Authorize(ctx context.Context, pr policies.Policy) error
}

// NewCallback creates a new instance of CallBack.
func NewCallback(httpClient *http.Client, method string, urls []string) CallBack {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &callback{
		httpClient: httpClient,
		urls:       urls,
		method:     method,
	}
}

func (c *callback) Authorize(ctx context.Context, pr policies.Policy) error {
	payload := map[string]string{
		"domain":           pr.Domain,
		"subject":          pr.Subject,
		"subject_type":     pr.SubjectType,
		"subject_kind":     pr.SubjectKind,
		"subject_relation": pr.SubjectRelation,
		"object":           pr.Object,
		"object_type":      pr.ObjectType,
		"object_kind":      pr.ObjectKind,
		"relation":         pr.Relation,
		"permission":       pr.Permission,
	}

	if len(c.urls) == 0 {
		return nil
	}

	if len(c.urls) == 1 {
		return c.makeRequest(ctx, c.method, c.urls[0], payload)
	}

	g, ctx := errgroup.WithContext(ctx)
	for i := range c.urls {
		url := c.urls[i]
		g.Go(func() error {
			return c.makeRequest(ctx, c.method, url, payload)
		})
	}

	return g.Wait()
}

func (c *callback) makeRequest(ctx context.Context, method, urlStr string, params map[string]string) error {
	var req *http.Request
	var err error

	if method == http.MethodGet {
		query := url.Values{}
		for key, value := range params {
			query.Set(key, value)
		}
		req, err = http.NewRequestWithContext(ctx, method, urlStr+"?"+query.Encode(), nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		data, jsonErr := json.Marshal(params)
		if jsonErr != nil {
			return jsonErr
		}
		req, err = http.NewRequestWithContext(ctx, method, urlStr, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
	}

	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Wrap(svcerr.ErrAuthorization, fmt.Errorf("status code %d", resp.StatusCode))
	}

	return nil
}
