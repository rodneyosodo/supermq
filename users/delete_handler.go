// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package users

import (
	"context"
	"log/slog"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/auth"
	mgclients "github.com/absmach/magistrala/pkg/clients"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/users/postgres"
)

const defLimit = uint64(100)

type handler struct {
	clients       postgres.Repository
	auth          magistrala.AuthServiceClient
	checkInterval time.Duration
	deleteAfter   time.Duration
	logger        *slog.Logger
}

func NewDeleteHandler(ctx context.Context, clients postgres.Repository, auth magistrala.AuthServiceClient, defCheckInterval, deleteAfter time.Duration, logger *slog.Logger) {
	handler := &handler{
		clients:       clients,
		auth:          auth,
		checkInterval: defCheckInterval,
		deleteAfter:   deleteAfter,
		logger:        logger,
	}

	go func() {
		ticker := time.NewTicker(handler.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				handler.routine(ctx)
			}
		}
	}()
}

func (h *handler) routine(ctx context.Context) {
	pm := mgclients.Page{Limit: defLimit, Offset: 0, Status: mgclients.DeletedStatus}

	dbUsers, err := h.clients.RetrieveAll(ctx, pm)
	if err != nil {
		h.logger.Error("failed to retrieve users", slog.Any("error", err))
		return
	}

	if dbUsers.Total > defLimit {
		for i := defLimit; i < dbUsers.Total; i += defLimit {
			pm.Offset = i
			du, err := h.clients.RetrieveAll(ctx, pm)
			if err != nil {
				h.logger.Error("failed to retrieve users", slog.Any("error", err), slog.Any("page", pm))
				return
			}

			dbUsers.Clients = append(dbUsers.Clients, du.Clients...)
		}
	}

	for _, u := range dbUsers.Clients {
		if time.Since(u.UpdatedAt) < h.deleteAfter {
			continue
		}

		deleteRes, err := h.auth.DeleteEntityPolicies(ctx, &magistrala.DeleteEntityPoliciesReq{
			Id:         u.ID,
			EntityType: auth.UserType,
		})
		if err != nil {
			h.logger.Error("failed to delete user policies", slog.Any("error", err))
			continue
		}
		if !deleteRes.Deleted {
			h.logger.Error("failed to delete user policies", slog.Any("error", svcerr.ErrAuthorization))
			continue
		}

		if err := h.clients.Delete(ctx, u.ID); err != nil {
			h.logger.Error("failed to delete user", slog.Any("error", err))
			continue
		}

		h.logger.Info("user deleted", slog.Group("user",
			slog.String("id", u.ID),
			slog.String("name", u.Name),
		))
	}
}
