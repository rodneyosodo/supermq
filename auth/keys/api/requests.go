// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"time"

	"github.com/mainflux/mainflux/auth/keys"
	"github.com/mainflux/mainflux/internal/apiutil"
)

type issueKeyReq struct {
	token    string
	Type     uint32        `json:"type,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

// It is not possible to issue Reset key using HTTP API.
func (req issueKeyReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.Type != keys.LoginKey &&
		req.Type != keys.RecoveryKey &&
		req.Type != keys.APIKey {
		return apiutil.ErrInvalidAPIKey
	}

	return nil
}

type keyReq struct {
	token string
	id    string
}

func (req keyReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}
