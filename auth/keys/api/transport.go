// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux/auth/keys"
	"github.com/mainflux/mainflux/internal/api"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc keys.Service, mux *bone.Mux, logger logger.Logger) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}
	mux.Post("/keys", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("issue"))(issueEndpoint(svc)),
		decodeIssue,
		api.EncodeResponse,
		opts...,
	))

	mux.Get("/keys/:id", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("retrieve"))(retrieveEndpoint(svc)),
		decodeKeyReq,
		api.EncodeResponse,
		opts...,
	))

	mux.Delete("/keys/:id", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("revoke"))(revokeEndpoint(svc)),
		decodeKeyReq,
		api.EncodeResponse,
		opts...,
	))
}

func decodeIssue(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	req := issueKeyReq{token: apiutil.ExtractBearerToken(r)}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeKeyReq(_ context.Context, r *http.Request) (interface{}, error) {
	req := keyReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, "id"),
	}
	return req, nil
}
