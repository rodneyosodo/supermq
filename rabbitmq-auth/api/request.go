// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"errors"

	apiutil "github.com/absmach/supermq/api/http/util"
)

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Vhost    string `json:"vhost"`
}

func (r *authRequest) Validate() error {
	if r.Username == "" {
		return apiutil.ErrMissingUsername
	}
	if r.Vhost == "" {
		return errors.New("missing vhost")
	}

	return nil
}
