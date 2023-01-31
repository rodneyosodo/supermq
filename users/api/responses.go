package api

import (
	"fmt"
	"net/http"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/users"
)

var (
	_ mainflux.Response = (*tokenRes)(nil)
	_ mainflux.Response = (*viewUserRes)(nil)
	_ mainflux.Response = (*createUserRes)(nil)
	_ mainflux.Response = (*updateUserRes)(nil)
	_ mainflux.Response = (*deleteUserRes)(nil)
	_ mainflux.Response = (*userssPageRes)(nil)
	_ mainflux.Response = (*viewMembersRes)(nil)
	_ mainflux.Response = (*memberPageRes)(nil)
)

// MailSent message response when link is sent
const MailSent = "Email with reset link is sent"

type pageRes struct {
	Limit  uint64 `json:"limit,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
	Total  uint64 `json:"total"`
}

type createUserRes struct {
	users.User
	created bool
}

func (res createUserRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res createUserRes) Headers() map[string]string {
	if res.created {
		return map[string]string{
			"Location": fmt.Sprintf("/users/%s", res.ID),
		}
	}

	return map[string]string{}
}

func (res createUserRes) Empty() bool {
	return false
}

type tokenRes struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	AccessType   string `json:"access_type,omitempty"`
}

func (res tokenRes) Code() int {
	return http.StatusCreated
}

func (res tokenRes) Headers() map[string]string {
	return map[string]string{}
}

func (res tokenRes) Empty() bool {
	return res.AccessToken == "" || res.RefreshToken == ""
}

type updateUserRes struct {
	users.User
}

func (res updateUserRes) Code() int {
	return http.StatusOK
}

func (res updateUserRes) Headers() map[string]string {
	return map[string]string{}
}

func (res updateUserRes) Empty() bool {
	return false
}

type viewUserRes struct {
	users.User
}

func (res viewUserRes) Code() int {
	return http.StatusOK
}

func (res viewUserRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewUserRes) Empty() bool {
	return false
}

type userssPageRes struct {
	pageRes
	Users []viewUserRes `json:"users"`
}

func (res userssPageRes) Code() int {
	return http.StatusOK
}

func (res userssPageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res userssPageRes) Empty() bool {
	return false
}

type viewMembersRes struct {
	users.User
}

func (res viewMembersRes) Code() int {
	return http.StatusOK
}

func (res viewMembersRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewMembersRes) Empty() bool {
	return false
}

type memberPageRes struct {
	pageRes
	Members []viewMembersRes `json:"members"`
}

func (res memberPageRes) Code() int {
	return http.StatusOK
}

func (res memberPageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res memberPageRes) Empty() bool {
	return false
}

type deleteUserRes struct {
	users.User
}

func (res deleteUserRes) Code() int {
	return http.StatusOK
}

func (res deleteUserRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deleteUserRes) Empty() bool {
	return false
}

type passwResetReqRes struct {
	Msg string `json:"msg"`
}

func (res passwResetReqRes) Code() int {
	return http.StatusCreated
}

func (res passwResetReqRes) Headers() map[string]string {
	return map[string]string{}
}

func (res passwResetReqRes) Empty() bool {
	return false
}

type passwChangeRes struct {
}

func (res passwChangeRes) Code() int {
	return http.StatusCreated
}

func (res passwChangeRes) Headers() map[string]string {
	return map[string]string{}
}

func (res passwChangeRes) Empty() bool {
	return false
}
