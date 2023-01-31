package api

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/users"
)

var _ users.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     users.Service
}

// MetricsMiddleware returns a new metrics middleware wrapper.
func MetricsMiddleware(svc users.Service, counter metrics.Counter, latency metrics.Histogram) users.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (ms *metricsMiddleware) Register(ctx context.Context, token string, user users.User) (users.User, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "register_user").Add(1)
		ms.latency.With("method", "register_user").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.Register(ctx, token, user)
}

func (ms *metricsMiddleware) Login(ctx context.Context, user users.User) (string, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "login").Add(1)
		ms.latency.With("method", "login").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.Login(ctx, user)
}

func (ms *metricsMiddleware) ViewUser(ctx context.Context, token, id string) (users.User, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_user").Add(1)
		ms.latency.With("method", "view_user").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ViewUser(ctx, token, id)
}

func (ms *metricsMiddleware) ViewProfile(ctx context.Context, token string) (users.User, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_profile").Add(1)
		ms.latency.With("method", "view_profile").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ViewProfile(ctx, token)
}

func (ms *metricsMiddleware) ListUsers(ctx context.Context, token string, pm users.Page) (users.UsersPage, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_users").Add(1)
		ms.latency.With("method", "list_users").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ListUsers(ctx, token, pm)
}

func (ms *metricsMiddleware) UpdateUser(ctx context.Context, token string, user users.User) (users.User, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_user_name_and_metadata").Add(1)
		ms.latency.With("method", "update_user_name_and_metadata").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateUser(ctx, token, user)
}

func (ms *metricsMiddleware) UpdateUserTags(ctx context.Context, token string, user users.User) (users.User, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_user_tags").Add(1)
		ms.latency.With("method", "update_user_tags").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateUserTags(ctx, token, user)
}

func (ms *metricsMiddleware) UpdateUserIdentity(ctx context.Context, token, id, identity string) (users.User, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_user_identity").Add(1)
		ms.latency.With("method", "update_user_identity").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateUserIdentity(ctx, token, id, identity)
}

func (ms *metricsMiddleware) ChangePassword(ctx context.Context, token, oldSecret, newSecret string) (users.User, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_user_password").Add(1)
		ms.latency.With("method", "update_user_password").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ChangePassword(ctx, token, oldSecret, newSecret)
}

func (ms *metricsMiddleware) UpdateUserOwner(ctx context.Context, token string, user users.User) (users.User, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_user_owner").Add(1)
		ms.latency.With("method", "update_user_owner").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateUserOwner(ctx, token, user)
}

func (ms *metricsMiddleware) GenerateResetToken(ctx context.Context, email, host string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "generate_reset_token").Add(1)
		ms.latency.With("method", "generate_reset_token").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.GenerateResetToken(ctx, email, host)
}

func (ms *metricsMiddleware) ResetPassword(ctx context.Context, token, password string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "reset_password").Add(1)
		ms.latency.With("method", "reset_password").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ResetPassword(ctx, token, password)
}

func (ms *metricsMiddleware) SendPasswordReset(ctx context.Context, host, email, token string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_user_owner").Add(1)
		ms.latency.With("method", "update_user_owner").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.SendPasswordReset(ctx, host, email, token)
}

func (ms *metricsMiddleware) EnableUser(ctx context.Context, token string, id string) (users.User, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "enable_user").Add(1)
		ms.latency.With("method", "enable_user").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.EnableUser(ctx, token, id)
}

func (ms *metricsMiddleware) DisableUser(ctx context.Context, token string, id string) (users.User, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "disable_user").Add(1)
		ms.latency.With("method", "disable_user").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.DisableUser(ctx, token, id)
}

func (ms *metricsMiddleware) ListMembers(ctx context.Context, token, groupID string, pm users.Page) (mp users.MembersPage, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_members").Add(1)
		ms.latency.With("method", "list_members").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ListMembers(ctx, token, groupID, pm)
}
