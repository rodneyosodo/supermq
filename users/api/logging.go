package api

import (
	"context"
	"fmt"
	"time"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/users"
)

var _ users.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    users.Service
}

func LoggingMiddleware(svc users.Service, logger log.Logger) users.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) Register(ctx context.Context, token string, user users.User) (u users.User, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method register_user of identity %s with token %s took %s to complete", user.Credentials.Identity, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.Register(ctx, token, user)
}

func (lm *loggingMiddleware) Login(ctx context.Context, user users.User) (token string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method issue_token for user %s took %s to complete", user.Credentials.Identity, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.Login(ctx, user)
}

func (lm *loggingMiddleware) ViewUser(ctx context.Context, token, id string) (u users.User, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view_user for user %s took %s to complete", u.Credentials.Identity, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ViewUser(ctx, token, id)
}

func (lm *loggingMiddleware) ViewProfile(ctx context.Context, token string) (u users.User, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view_user for user %s took %s to complete", u.Credentials.Identity, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ViewProfile(ctx, token)
}

func (lm *loggingMiddleware) ListUsers(ctx context.Context, token string, pm users.Page) (cp users.UsersPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_users for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ListUsers(ctx, token, pm)
}

func (lm *loggingMiddleware) UpdateUser(ctx context.Context, token string, user users.User) (u users.User, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_user_name_and_metadata for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateUser(ctx, token, user)
}

func (lm *loggingMiddleware) UpdateUserTags(ctx context.Context, token string, user users.User) (u users.User, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_user_tags for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateUserTags(ctx, token, user)
}

func (lm *loggingMiddleware) UpdateUserIdentity(ctx context.Context, token, id, identity string) (u users.User, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_user_identity for token %s and identity %s took %s to complete", token, identity, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateUserIdentity(ctx, token, id, identity)
}

func (lm *loggingMiddleware) ChangePassword(ctx context.Context, token, oldSecret, newSecret string) (u users.User, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_user_secret for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ChangePassword(ctx, token, oldSecret, newSecret)
}

func (lm *loggingMiddleware) UpdateUserOwner(ctx context.Context, token string, user users.User) (u users.User, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_user_owner for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateUserOwner(ctx, token, user)
}

func (lm *loggingMiddleware) GenerateResetToken(ctx context.Context, email, host string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method generate_reset_token for email %s took %s to complete", email, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.GenerateResetToken(ctx, email, host)
}

func (lm *loggingMiddleware) ResetPassword(ctx context.Context, token, password string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method reset_password for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ResetPassword(ctx, token, password)
}

func (lm *loggingMiddleware) SendPasswordReset(ctx context.Context, email, host, token string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_user_owner for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.SendPasswordReset(ctx, email, host, token)
}

func (lm *loggingMiddleware) EnableUser(ctx context.Context, token string, id string) (u users.User, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method enable_user for user %s took %s to complete", id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.EnableUser(ctx, token, id)
}

func (lm *loggingMiddleware) DisableUser(ctx context.Context, token string, id string) (u users.User, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method disable_user for user %s took %s to complete", id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.DisableUser(ctx, token, id)
}

func (lm *loggingMiddleware) ListMembers(ctx context.Context, token, groupID string, cp users.Page) (mp users.MembersPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_members for group %s and token %s took %s to complete", groupID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ListMembers(ctx, token, groupID, cp)
}
