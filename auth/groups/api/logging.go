package api

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux/auth/groups"
	log "github.com/mainflux/mainflux/logger"
)

var _ groups.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    groups.Service
}

func LoggingMiddleware(svc groups.Service, logger log.Logger) groups.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) CreateGroup(ctx context.Context, token string, group groups.Group) (rGroup groups.Group, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method create_group for group %s and token %s took %s to complete", group.Name, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.CreateGroup(ctx, token, group)
}

func (lm *loggingMiddleware) UpdateGroup(ctx context.Context, token string, group groups.Group) (rGroup groups.Group, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_group for group %s and token %s took %s to complete", group.Name, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateGroup(ctx, token, group)
}

func (lm *loggingMiddleware) ViewGroup(ctx context.Context, token, id string) (g groups.Group, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view_group for group %s and token %s took %s to complete", g.Name, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ViewGroup(ctx, token, id)
}

func (lm *loggingMiddleware) ListGroups(ctx context.Context, token string, gp groups.GroupsPage) (cg groups.GroupsPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_groups for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ListGroups(ctx, token, gp)
}

func (lm *loggingMiddleware) EnableGroup(ctx context.Context, token string, id string) (g groups.Group, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method enable_group for client %s took %s to complete", id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.EnableGroup(ctx, token, id)
}

func (lm *loggingMiddleware) DisableGroup(ctx context.Context, token string, id string) (g groups.Group, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method disable_group for client %s took %s to complete", id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.DisableGroup(ctx, token, id)
}

func (lm *loggingMiddleware) ListMemberships(ctx context.Context, token, clientID string, cp groups.GroupsPage) (mp groups.MembershipsPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_memberships for group %s and token %s took %s to complete", clientID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ListMemberships(ctx, token, clientID, cp)
}
