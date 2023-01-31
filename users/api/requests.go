package api

import (
	"github.com/mainflux/mainflux/internal/api"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/users"
)

const (
	maxLimitSize = 100
	maxEmailSize = 1024
)

type createUserReq struct {
	user  users.User
	token string
}

func (req createUserReq) validate() error {
	if len(req.user.Name) > api.MaxNameSize {
		return apiutil.ErrNameSize
	}

	return req.user.Validate()
}

type viewUserReq struct {
	token string
	id    string
}

func (req viewUserReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	return nil
}

type listUsersReq struct {
	token      string
	status     users.Status
	offset     uint64
	limit      uint64
	name       string
	tag        string
	owner      string
	sharedBy   string
	visibility string
	metadata   users.Metadata
	email      string
}

func (req listUsersReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.limit > maxLimitSize || req.limit < 1 {
		return apiutil.ErrLimitSize
	}
	if len(req.email) > maxEmailSize {
		return apiutil.ErrEmailSize
	}
	if req.visibility != "" &&
		req.visibility != api.AllVisibility &&
		req.visibility != api.MyVisibility &&
		req.visibility != api.SharedVisibility {
		return apiutil.ErrInvalidVisibilityType
	}
	if req.status != users.AllStatus &&
		req.status != users.EnabledStatus &&
		req.status != users.DisabledStatus {
		return apiutil.ErrInvalidStatus
	}
	return nil
}

type listMembersReq struct {
	users.Page
	token   string
	groupID string
}

func (req listMembersReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.groupID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type updateUserReq struct {
	token    string
	id       string
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateUserReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type updateUserTagsReq struct {
	id    string
	token string
	Tags  []string `json:"tags,omitempty"`
}

func (req updateUserTagsReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	if len(req.Tags) == 0 {
		return apiutil.ErrMissingUserTags
	}
	return nil
}

type updateUserOwnerReq struct {
	id    string
	token string
	Owner string `json:"owner,omitempty"`
}

func (req updateUserOwnerReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	if req.Owner == "" {
		return apiutil.ErrMissingOwner
	}
	return nil
}

type updateUserCredentialsReq struct {
	token     string
	id        string
	Identity  string `json:"email,omitempty"`
	OldSecret string `json:"old_secret,omitempty"`
	NewSecret string `json:"new_secret,omitempty"`
}

func (req updateUserCredentialsReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	return nil
}

type passwResetReq struct {
	Email string `json:"email"`
	Host  string `json:"host"`
}

func (req passwResetReq) validate() error {
	if req.Email == "" {
		return apiutil.ErrMissingEmail
	}

	if req.Host == "" {
		return apiutil.ErrMissingHost
	}

	return nil
}

type resetTokenReq struct {
	Token    string `json:"token"`
	Password string `json:"password"`
	ConfPass string `json:"confirm_password"`
}

func (req resetTokenReq) validate() error {
	if req.Password == "" {
		return apiutil.ErrMissingPass
	}

	if req.ConfPass == "" {
		return apiutil.ErrMissingConfPass
	}

	if req.Token == "" {
		return apiutil.ErrBearerToken
	}

	if req.Password != req.ConfPass {
		return apiutil.ErrInvalidResetPass
	}

	return nil
}

type changeUserStatusReq struct {
	token string
	id    string
}

func (req changeUserStatusReq) validate() error {
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type loginUserReq struct {
	Identity string `json:"email,omitempty"`
	Secret   string `json:"password,omitempty"`
}

func (req loginUserReq) validate() error {
	if req.Identity == "" {
		return apiutil.ErrMissingIdentity
	}
	if req.Secret == "" {
		return apiutil.ErrMissingSecret
	}
	return nil
}

type tokenReq struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

func (req tokenReq) validate() error {
	if req.RefreshToken == "" {
		return apiutil.ErrBearerToken
	}
	return nil
}
