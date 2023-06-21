// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc

type authorizeRes struct {
	authorized bool
}

type identifyRes struct {
	id string
}

type issueRes struct {
	token string
}

type addPolicyRes struct {
	added bool
}

type deletePolicyRes struct {
	deleted bool
}
