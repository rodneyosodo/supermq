// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
)

const certsEndpoint = "certs"

// Cert represents certs data.
type Cert struct {
	ThingID    string    `json:"thing_id,omitempty"`
	CertSerial string    `json:"cert_serial,omitempty"`
	ClientKey  string    `json:"client_key,omitempty"`
	ClientCert string    `json:"client_cert,omitempty"`
	Expiration time.Time `json:"expiration,omitempty"`
}

func (sdk mfSDK) IssueCert(thingID string, keyBits int, keyType, valid, token string) (Cert, errors.SDKError) {
	r := certReq{
		ThingID: thingID,
		KeyBits: keyBits,
		KeyType: keyType,
		Valid:   valid,
	}
	d, err := json.Marshal(r)
	if err != nil {
		return Cert{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s", sdk.certsURL, certsEndpoint)
	_, body, sdkerr := sdk.processRequest(http.MethodPost, url, token, string(CTJSON), d, http.StatusCreated)
	if sdkerr != nil {
		return Cert{}, sdkerr
	}

	var c Cert
	if err := json.Unmarshal(body, &c); err != nil {
		return Cert{}, errors.NewSDKError(err)
	}
	return c, nil
}

func (sdk mfSDK) ViewCert(id, token string) (Cert, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s", sdk.certsURL, certsEndpoint, id)
	fmt.Println(url)
	_, body, err := sdk.processRequest(http.MethodGet, url, token, string(CTJSON), nil, http.StatusOK)
	if err != nil {
		return Cert{}, err
	}

	var cert Cert
	if err := json.Unmarshal(body, &cert); err != nil {
		return Cert{}, errors.NewSDKError(err)
	}

	return cert, nil
}

func (sdk mfSDK) RevokeCert(id, token string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s", sdk.certsURL, certsEndpoint, id)
	_, _, err := sdk.processRequest(http.MethodDelete, url, token, string(CTJSON), nil, http.StatusOK)
	return err
}

type certReq struct {
	ThingID    string `json:"thing_id"`
	KeyBits    int    `json:"key_bits"`
	KeyType    string `json:"key_type"`
	Encryption string `json:"encryption"`
	Valid      string `json:"ttl"`
}
