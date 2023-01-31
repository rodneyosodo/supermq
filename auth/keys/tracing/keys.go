// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package tracing contains middlewares that will add spans
// to existing traces.
package tracing

import (
	"context"

	"github.com/mainflux/mainflux/auth/keys"
	"go.opentelemetry.io/otel/trace"
)

const (
	saveOp     = "save"
	retrieveOp = "retrieve_by_id"
	revokeOp   = "remove"
)

var _ keys.Service = (*keyTracingMiddleware)(nil)

// keyTracingMiddleware tracks request and their latency, and adds spans
// to context.
type keyTracingMiddleware struct {
	tracer trace.Tracer
	svc    keys.Service
}

// TracingMiddleware tracks request and their latency, and adds spans
// to context.
func TracingMiddleware(svc keys.Service, tracer trace.Tracer) keys.Service {
	return keyTracingMiddleware{
		tracer: tracer,
		svc:    svc,
	}
}

func (krm keyTracingMiddleware) Issue(ctx context.Context, token string, key keys.Key) (keys.Key, string, error) {
	ctx, span := krm.tracer.Start(ctx, "svc_remove_key")
	defer span.End()
	return krm.svc.Issue(ctx, token, key)
}

func (krm keyTracingMiddleware) Revoke(ctx context.Context, token, id string) error {
	ctx, span := krm.tracer.Start(ctx, "svc_revoke_key")
	defer span.End()
	return krm.svc.Revoke(ctx, token, id)
}

func (krm keyTracingMiddleware) RetrieveKey(ctx context.Context, token, id string) (keys.Key, error) {
	ctx, span := krm.tracer.Start(ctx, "svc_retrieve_key")
	defer span.End()
	return krm.svc.RetrieveKey(ctx, token, id)
}

func (krm keyTracingMiddleware) Identify(ctx context.Context, token string) (keys.Identity, error) {
	ctx, span := krm.tracer.Start(ctx, "svc_identify")
	defer span.End()
	return krm.svc.Identify(ctx, token)
}
