// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"

	"github.com/absmach/supermq"
)

var _ supermq.Response = (*authResponse)(nil)

type authResponse struct {
	authenticated bool
}

func (i authResponse) Code() int {
	if i.authenticated {
		return http.StatusOK
	}

	return http.StatusUnauthorized
}

func (i authResponse) Headers() map[string]string {
	return map[string]string{}
}

func (i authResponse) Empty() bool {
	return true
}
