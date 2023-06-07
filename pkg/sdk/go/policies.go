package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	policiesEndpoint  = "policies"
	authorizeEndpoint = "authorize"
	accessEndpoint    = "access"
)

var errInvalidAction = errors.NewSDKError(fmt.Errorf("invalid number of actions, it should be exactly 1"))

// Policy represents an argument struct for making a policy related function calls.
type Policy struct {
	OwnerID   string    `json:"owner_id"`
	Subject   string    `json:"subject"`
	Object    string    `json:"object"`
	Actions   []string  `json:"actions"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (sdk mfSDK) Authorize(p Policy, entityType, token string) (bool, errors.SDKError) {
	if len(p.Actions) == 1 {
		return false, errInvalidAction
	}
	areq := authorizeReq{Subject: p.Subject, Object: p.Object, Action: p.Actions[0], EntityType: entityType}
	data, err := json.Marshal(areq)
	if err != nil {
		return false, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s", sdk.usersURL, authorizeEndpoint)
	_, body, sdkerr := sdk.processRequest(http.MethodPost, url, token, string(CTJSON), data, http.StatusOK)
	if sdkerr != nil {
		return false, sdkerr
	}
	resp := authorizeRes{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return false, errors.NewSDKError(err)
	}

	return resp.Authorized, nil
}

func (sdk mfSDK) CreatePolicy(p Policy, token string) errors.SDKError {
	data, err := json.Marshal(p)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s", sdk.usersURL, policiesEndpoint)
	_, _, sdkerr := sdk.processRequest(http.MethodPost, url, token, string(CTJSON), data, http.StatusCreated)
	if sdkerr != nil {
		return sdkerr
	}

	return nil
}

func (sdk mfSDK) UpdatePolicy(p Policy, token string) errors.SDKError {
	data, err := json.Marshal(p)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s", sdk.usersURL, policiesEndpoint)

	_, _, sdkerr := sdk.processRequest(http.MethodPut, url, token, string(CTJSON), data, http.StatusNoContent)
	if sdkerr != nil {
		return sdkerr
	}

	return nil
}

func (sdk mfSDK) ListPolicies(pm PageMetadata, token string) (PolicyPage, errors.SDKError) {
	url, err := sdk.withQueryParams(sdk.usersURL, policiesEndpoint, pm)
	if err != nil {
		return PolicyPage{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, string(CTJSON), nil, http.StatusOK)
	if sdkerr != nil {
		return PolicyPage{}, sdkerr
	}

	var pp PolicyPage
	if err := json.Unmarshal(body, &pp); err != nil {
		return PolicyPage{}, errors.NewSDKError(err)
	}

	return pp, nil
}

func (sdk mfSDK) DeletePolicy(p Policy, token string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s/%s", sdk.usersURL, policiesEndpoint, p.Subject, p.Object)

	_, _, sdkerr := sdk.processRequest(http.MethodDelete, url, token, string(CTJSON), nil, http.StatusNoContent)
	return sdkerr
}

func (sdk mfSDK) Assign(memberType []string, memberID, groupID, token string) errors.SDKError {
	var policy = Policy{
		Subject: memberID,
		Object:  groupID,
		Actions: memberType,
	}
	data, err := json.Marshal(policy)
	if err != nil {
		return errors.NewSDKError(err)
	}
	url := fmt.Sprintf("%s/%s", sdk.usersURL, policiesEndpoint)
	_, _, sdkerr := sdk.processRequest(http.MethodPost, url, token, string(CTJSON), data, http.StatusCreated)
	return sdkerr
}

func (sdk mfSDK) Unassign(memberType []string, groupID string, memberID string, token string) errors.SDKError {
	var policy = Policy{
		Subject: memberID,
		Object:  groupID,
		Actions: memberType,
	}
	data, err := json.Marshal(policy)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/%s", sdk.usersURL, policiesEndpoint, groupID, memberID)

	_, _, sdkerr := sdk.processRequest(http.MethodDelete, url, token, string(CTJSON), data, http.StatusNoContent)
	return sdkerr
}

func (sdk mfSDK) Connect(connIDs ConnectionIDs, token string) errors.SDKError {
	data, err := json.Marshal(connIDs)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s", sdk.thingsURL, connectEndpoint)

	_, _, sdkerr := sdk.processRequest(http.MethodPost, url, token, string(CTJSON), data, http.StatusCreated)
	return sdkerr
}

func (sdk mfSDK) Disconnect(connIDs ConnectionIDs, token string) errors.SDKError {
	data, err := json.Marshal(connIDs)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s", sdk.thingsURL, disconnectEndpoint)
	_, _, sdkerr := sdk.processRequest(http.MethodPost, url, token, string(CTJSON), data, http.StatusNoContent)
	return sdkerr
}

func (sdk mfSDK) ConnectThing(thingID, chanID, token string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", sdk.thingsURL, channelsEndpoint, chanID, thingsEndpoint, thingID)

	_, _, err := sdk.processRequest(http.MethodPost, url, token, string(CTJSON), nil, http.StatusCreated)
	return err
}

func (sdk mfSDK) DisconnectThing(thingID, chanID, token string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", sdk.thingsURL, channelsEndpoint, chanID, thingsEndpoint, thingID)

	_, _, err := sdk.processRequest(http.MethodDelete, url, token, string(CTJSON), nil, http.StatusNoContent)
	return err
}

func (sdk mfSDK) UpdateThingsPolicy(p Policy, token string) errors.SDKError {
	data, err := json.Marshal(p)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s", sdk.thingsURL, thingsEndpoint, policiesEndpoint)

	_, _, sdkerr := sdk.processRequest(http.MethodPut, url, token, string(CTJSON), data, http.StatusNoContent)
	if sdkerr != nil {
		return sdkerr
	}

	return nil
}

func (sdk mfSDK) ListThingsPolicies(pm PageMetadata, token string) (PolicyPage, errors.SDKError) {
	url, err := sdk.withQueryParams(fmt.Sprintf("%s/%s", sdk.thingsURL, thingsEndpoint), policiesEndpoint, pm)
	if err != nil {
		return PolicyPage{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, string(CTJSON), nil, http.StatusOK)
	if sdkerr != nil {
		return PolicyPage{}, sdkerr
	}

	var pp PolicyPage
	if err := json.Unmarshal(body, &pp); err != nil {
		return PolicyPage{}, errors.NewSDKError(err)
	}

	return pp, nil
}

func (sdk mfSDK) ThingCanAccess(p Policy, entityType, token string) (bool, string, errors.SDKError) {
	if len(p.Actions) != 1 {
		return false, "", errInvalidAction
	}
	creq := canAccessReq{ClientSecret: p.Subject, Action: p.Actions[0], EntityType: entityType}
	data, err := json.Marshal(creq)
	if err != nil {
		return false, "", errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/%s", sdk.thingsURL, channelsEndpoint, p.Object, accessEndpoint)
	_, body, sdkerr := sdk.processRequest(http.MethodPost, url, token, string(CTJSON), data, http.StatusOK)
	if sdkerr != nil {
		return false, "", sdkerr
	}
	resp := canAccessRes{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return false, "", errors.NewSDKError(err)
	}

	return resp.Authorized, resp.ThingID, nil
}
