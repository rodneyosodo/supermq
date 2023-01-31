package grpc

import (
	"github.com/mainflux/mainflux/auth/keys"
	"github.com/mainflux/mainflux/internal/apiutil"
)

type identityReq struct {
	token string
	kind  uint32
}

func (req identityReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.kind != keys.LoginKey &&
		req.kind != keys.APIKey &&
		req.kind != keys.RecoveryKey {
		return apiutil.ErrInvalidAuthKey
	}

	return nil
}

type issueReq struct {
	id      string
	email   string
	keyType uint32
}

func (req issueReq) validate() error {
	if req.email == "" {
		return apiutil.ErrMissingEmail
	}
	if req.keyType != keys.LoginKey &&
		req.keyType != keys.APIKey &&
		req.keyType != keys.RecoveryKey {
		return apiutil.ErrInvalidAuthKey
	}

	return nil
}

// authReq represents authorization request. It contains:
// 1. subject - an action invoker (client)
// 2. object - an entity over which action will be executed (client, group, computation, dataset)
// 3. action - type of action that will be executed (read/write)
type authReq struct {
	Sub        string
	Obj        string
	Act        string
	EntityType string
}

func (req authReq) validate() error {
	if req.Sub == "" {
		return apiutil.ErrMissingPolicySub
	}
	if req.Obj == "" {
		return apiutil.ErrMissingPolicyObj
	}
	if req.Act == "" {
		return apiutil.ErrMissingPolicyAct
	}
	if req.EntityType == "" {
		return apiutil.ErrMissingPolicyEntityType
	}

	return nil
}

type assignReq struct {
	token     string
	groupID   string
	memberID  string
	groupType string
}

func (req assignReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.groupID == "" || req.memberID == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type membersReq struct {
	token      string
	groupID    string
	offset     uint64
	limit      uint64
	memberType string
}

func (req membersReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.groupID == "" {
		return apiutil.ErrMissingID
	}
	if req.memberType == "" {
		return apiutil.ErrMissingMemberType
	}
	return nil
}

type policyReq struct {
	Sub string
	Obj string
	Act string
}

func (req policyReq) validate() error {
	if req.Sub == "" {
		return apiutil.ErrMissingPolicySub
	}

	if req.Obj == "" {
		return apiutil.ErrMissingPolicyObj
	}

	if req.Act == "" {
		return apiutil.ErrMissingPolicyAct
	}

	return nil
}

type listPoliciesReq struct {
	Sub string
	Obj string
	Act string
}
