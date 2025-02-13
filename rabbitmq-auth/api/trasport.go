// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/absmach/supermq"
	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/pkg/errors"
	rabbitmqauth "github.com/absmach/supermq/rabbitmq-auth"
	"github.com/go-chi/chi/v5"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func MakeHandler(svc rabbitmqauth.Service, logger *slog.Logger, instanceID string) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, encodeError)),
	}

	mux := chi.NewRouter()

	mux.Route("/auth", func(r chi.Router) {
		r.HandleFunc("/user", otelhttp.NewHandler(kithttp.NewServer(
			authenticateEndpoint(svc),
			decodeAuthenticateReq,
			encodeResponse,
			opts...,
		), "authenticate-user").ServeHTTP)
		r.HandleFunc("/vhost", otelhttp.NewHandler(kithttp.NewServer(
			authenticateEndpoint(svc),
			decodeAuthenticateReq,
			encodeResponse,
			opts...,
		), "authenticate-vhost").ServeHTTP)
		r.HandleFunc("/resource", otelhttp.NewHandler(kithttp.NewServer(
			authenticateEndpoint(svc),
			decodeAuthenticateReq,
			encodeResponse,
			opts...,
		), "authenticate-resource").ServeHTTP)
		r.HandleFunc("/topic", otelhttp.NewHandler(kithttp.NewServer(
			authenticateEndpoint(svc),
			decodeAuthenticateReq,
			encodeResponse,
			opts...,
		), "authenticate-topic").ServeHTTP)
	})

	mux.Get("/health", supermq.Health("cube-proxy", instanceID))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

func decodeAuthenticateReq(_ context.Context, r *http.Request) (interface{}, error) {
	switch r.Method {
	case http.MethodPost:
		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(err, errors.ErrMalformedEntity))
		}

		return req, nil
	case http.MethodGet:
		username, err := apiutil.ReadStringQuery(r, "username", "")
		if err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		password, err := apiutil.ReadStringQuery(r, "password", "")
		if err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		vhost, err := apiutil.ReadStringQuery(r, "vhost", "")
		if err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		return authRequest{Username: username, Password: password, Vhost: vhost}, nil
	default:
		return nil, errors.New("invalid method")
	}
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	if ar, ok := response.(supermq.Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}
		switch ar.Code() {
		case http.StatusOK:
			w.WriteHeader(ar.Code())
			if _, err := w.Write([]byte(`allow`)); err != nil {
				return err
			}
		default:
			w.WriteHeader(http.StatusUnauthorized)
			if _, err := w.Write([]byte(`deny`)); err != nil {
				return err
			}
		}

		return nil
	}

	if _, err := w.Write([]byte("deny")); err != nil {
		return err
	}

	return nil
}

func encodeError(_ context.Context, _ error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	if _, err := w.Write([]byte("deny")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
