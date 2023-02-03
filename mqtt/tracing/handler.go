package tracing

import (
	"context"

	"github.com/mainflux/mproxy/pkg/session"
	"go.opentelemetry.io/otel/trace"
)

const (
	authConnectOP   = "auth_connect_op"
	authPublishOP   = "auth_publish_op"
	authSubscribeOP = "auth_subscribe_op"
	connectOP       = "connect_op"
	disconnectOP    = "disconnect_op"
	subscribeOP     = "subscribe_op"
	unsubscribeOP   = "unsubscribe_op"
	publishOP       = "publish_op"
)

var _ session.Handler = (*handlerMiddleware)(nil)

type handlerMiddleware struct {
	handler session.Handler
	tracer  trace.Tracer
}

// NewHandler creates a new session.Handler middleware with tracing.
func NewHandler(tracer trace.Tracer, handler session.Handler) session.Handler {
	return &handlerMiddleware{
		tracer:  tracer,
		handler: handler,
	}
}

// AuthConnect traces auth connect operations.
func (h *handlerMiddleware) AuthConnect(ctx context.Context) error {
	// span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, h.tracer, authConnectOP)
	ctx, span := h.tracer.Start(ctx, authConnectOP)
	defer span.End()
	return h.handler.AuthConnect(ctx)
}

// AuthPublish traces auth publish operations.
func (h *handlerMiddleware) AuthPublish(ctx context.Context, topic *string, payload *[]byte) error {
	ctx, span := h.tracer.Start(ctx, authPublishOP)
	defer span.End()
	return h.handler.AuthPublish(ctx, topic, payload)
}

// AuthSubscribe traces auth subscribe operations.
func (h *handlerMiddleware) AuthSubscribe(ctx context.Context, topics *[]string) error {
	ctx, span := h.tracer.Start(ctx, authSubscribeOP)
	defer span.End()
	return h.handler.AuthSubscribe(ctx, topics)
}

// Connect traces connect operations.
func (h *handlerMiddleware) Connect(ctx context.Context) {
	ctx, span := h.tracer.Start(ctx, connectOP)
	defer span.End()
	h.handler.Connect(ctx)
}

// Disconnect traces disconnect operations.
func (h *handlerMiddleware) Disconnect(ctx context.Context) {
	ctx, span := h.tracer.Start(ctx, disconnectOP)
	defer span.End()
	h.handler.Disconnect(ctx)
}

// Publish traces publish operations.
func (h *handlerMiddleware) Publish(ctx context.Context, topic *string, payload *[]byte) {
	ctx, span := h.tracer.Start(ctx, publishOP)
	defer span.End()
	h.handler.Publish(ctx, topic, payload)
}

// Subscribe traces subscribe operations.
func (h *handlerMiddleware) Subscribe(ctx context.Context, topics *[]string) {
	ctx, span := h.tracer.Start(ctx, subscribeOP)
	defer span.End()
	h.handler.Subscribe(ctx, topics)
}

// Unsubscribe traces unsubscribe operations.
func (h *handlerMiddleware) Unsubscribe(ctx context.Context, topics *[]string) {
	ctx, span := h.tracer.Start(ctx, unsubscribeOP)
	defer span.End()
	h.handler.Unsubscribe(ctx, topics)
}
