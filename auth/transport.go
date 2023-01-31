// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package http

import (
	"net/http"

	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/auth/groups"
	groupsapi "github.com/mainflux/mainflux/auth/groups/api"
	"github.com/mainflux/mainflux/auth/keys"
	keysapi "github.com/mainflux/mainflux/auth/keys/api"
	"github.com/mainflux/mainflux/auth/policies"
	policiesapi "github.com/mainflux/mainflux/auth/policies/api/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(gsvc groups.Service, ksvc keys.Service, psvc policies.Service, tracer opentracing.Tracer, logger logger.Logger) http.Handler {
	mux := bone.New()
	keysapi.MakeHandler(ksvc, mux, logger)
	groupsapi.MakeGroupsHandler(gsvc, mux, logger)
	policiesapi.MakePolicyHandler(psvc, mux, logger)
	mux.GetFunc("/health", mainflux.Health("auth"))
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}
