package tracing

import (
	"context"

	"github.com/mainflux/mainflux/users"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ users.Service = (*tracingMiddleware)(nil)

type tracingMiddleware struct {
	tracer trace.Tracer
	svc    users.Service
}

func TracingMiddleware(svc users.Service, tracer trace.Tracer) users.Service {
	return &tracingMiddleware{tracer, svc}
}

func (tm *tracingMiddleware) Register(ctx context.Context, token string, user users.User) (users.User, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_register_user", trace.WithAttributes(attribute.String("identity", user.Credentials.Identity)))
	defer span.End()

	return tm.svc.Register(ctx, token, user)
}

func (tm *tracingMiddleware) Login(ctx context.Context, user users.User) (string, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_issue_token", trace.WithAttributes(attribute.String("identity", user.Credentials.Identity)))
	defer span.End()

	return tm.svc.Login(ctx, user)
}

func (tm *tracingMiddleware) ViewUser(ctx context.Context, token string, id string) (users.User, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_view_user", trace.WithAttributes(attribute.String("ID", id)))
	defer span.End()
	return tm.svc.ViewUser(ctx, token, id)
}

func (tm *tracingMiddleware) ListUsers(ctx context.Context, token string, pm users.Page) (users.UsersPage, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_list_users")
	defer span.End()
	return tm.svc.ListUsers(ctx, token, pm)
}

func (tm *tracingMiddleware) UpdateUser(ctx context.Context, token string, user users.User) (users.User, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_user_name_and_metadata", trace.WithAttributes(attribute.String("Name", user.Name)))
	defer span.End()

	return tm.svc.UpdateUser(ctx, token, user)
}

func (tm *tracingMiddleware) UpdateUserTags(ctx context.Context, token string, user users.User) (users.User, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_user_tags", trace.WithAttributes(attribute.StringSlice("Tags", user.Tags)))
	defer span.End()

	return tm.svc.UpdateUserTags(ctx, token, user)
}
func (tm *tracingMiddleware) UpdateUserIdentity(ctx context.Context, token, id, identity string) (users.User, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_user_identity", trace.WithAttributes(attribute.String("Identity", identity)))
	defer span.End()

	return tm.svc.UpdateUserIdentity(ctx, token, id, identity)

}

func (tm *tracingMiddleware) ChangePassword(ctx context.Context, token, oldSecret, newSecret string) (users.User, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_user_password")
	defer span.End()

	return tm.svc.ChangePassword(ctx, token, oldSecret, newSecret)

}

func (tm *tracingMiddleware) ViewProfile(ctx context.Context, token string) (users.User, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_view_profile")
	defer span.End()

	return tm.svc.ViewProfile(ctx, token)

}

func (tm *tracingMiddleware) GenerateResetToken(ctx context.Context, email, host string) error {
	ctx, span := tm.tracer.Start(ctx, "svc_generate_reset_token")
	defer span.End()

	return tm.svc.GenerateResetToken(ctx, email, host)

}

func (tm *tracingMiddleware) ResetPassword(ctx context.Context, token, secret string) error {
	ctx, span := tm.tracer.Start(ctx, "svc_reset_password")
	defer span.End()

	return tm.svc.ResetPassword(ctx, token, secret)

}

func (tm *tracingMiddleware) SendPasswordReset(ctx context.Context, host, email, token string) error {
	ctx, span := tm.tracer.Start(ctx, "svc_update_user_password")
	defer span.End()

	return tm.svc.SendPasswordReset(ctx, host, email, token)

}

func (tm *tracingMiddleware) UpdateUserOwner(ctx context.Context, token string, user users.User) (users.User, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_user_owner", trace.WithAttributes(attribute.String("Owner", user.Owner)))
	defer span.End()

	return tm.svc.UpdateUserOwner(ctx, token, user)
}

func (tm *tracingMiddleware) EnableUser(ctx context.Context, token, id string) (users.User, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_enable_user", trace.WithAttributes(attribute.String("ID", id)))
	defer span.End()

	return tm.svc.EnableUser(ctx, token, id)
}

func (tm *tracingMiddleware) DisableUser(ctx context.Context, token, id string) (users.User, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_disable_user", trace.WithAttributes(attribute.String("ID", id)))
	defer span.End()

	return tm.svc.DisableUser(ctx, token, id)
}

func (tm *tracingMiddleware) ListMembers(ctx context.Context, token, groupID string, pm users.Page) (users.MembersPage, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_list_members")
	defer span.End()

	return tm.svc.ListMembers(ctx, token, groupID, pm)

}
