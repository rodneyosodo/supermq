// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package producer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mainflux/mainflux/bootstrap"
)

const (
	configPrefix        = "config."
	configCreate        = configPrefix + "create"
	configUpdate        = configPrefix + "update"
	configRemove        = configPrefix + "remove"
	configList          = configPrefix + "list"
	configHandlerRemove = configPrefix + "handler_remove"

	thingPrefix            = "thing."
	thingBootstrap         = thingPrefix + "bootstrap"
	thingStateChange       = thingPrefix + "state_change"
	thingUpdateConnections = thingPrefix + "update_connections"
	channelHandlerRemove   = thingPrefix + "channel_handler_remove"
	channelUpdateHandler   = thingPrefix + "update_channel_handler"
	thingDisconnect        = thingPrefix + "disconnect"

	certUpdate = "cert.update"
)

type event interface {
	encode() map[string]interface{}
}

var (
	_ event = (*configEvent)(nil)
	_ event = (*removeConfigEvent)(nil)
	_ event = (*bootstrapEvent)(nil)
	_ event = (*changeStateEvent)(nil)
	_ event = (*updateConnectionsEvent)(nil)
	_ event = (*updateCertEvent)(nil)
	_ event = (*listConfigsEvent)(nil)
	_ event = (*removeHandlerEvent)(nil)
)

type configEvent struct {
	bootstrap.Config
}

func (ce configEvent) encode() map[string]interface{} {
	val := map[string]interface{}{
		"state":     ce.State.String(),
		"operation": configCreate,
	}
	if ce.MFThing != "" {
		val["mainflux_thing"] = ce.MFThing
	}
	if ce.Content != "" {
		val["content"] = ce.Content
	}
	if ce.Owner != "" {
		val["owner"] = ce.Owner
	}
	if ce.Name != "" {
		val["name"] = ce.Name
	}
	if ce.ExternalID != "" {
		val["external_id"] = ce.ExternalID
	}
	if len(ce.MFChannels) > 0 {
		channels := make([]string, len(ce.MFChannels))
		for i, ch := range ce.MFChannels {
			channels[i] = ch.ID
		}
		val["channels"] = fmt.Sprintf("[%s]", strings.Join(channels, ", "))
	}
	if ce.ClientCert != "" {
		val["client_cert"] = ce.ClientCert
	}
	if ce.ClientKey != "" {
		val["client_key"] = ce.ClientKey
	}
	if ce.CACert != "" {
		val["ca_cert"] = ce.CACert
	}
	if ce.Content != "" {
		val["content"] = ce.Content
	}

	return val
}

type removeConfigEvent struct {
	mfThing string
}

func (rce removeConfigEvent) encode() map[string]interface{} {
	return map[string]interface{}{
		"thing_id":  rce.mfThing,
		"operation": configRemove,
	}
}

type listConfigsEvent struct {
	offset       uint64
	limit        uint64
	fullMatch    map[string]string
	partialMatch map[string]string
}

func (rce listConfigsEvent) encode() map[string]interface{} {
	val := map[string]interface{}{
		"offset":    rce.offset,
		"limit":     rce.limit,
		"operation": configList,
	}
	if len(rce.fullMatch) > 0 {
		data, err := json.Marshal(rce.fullMatch)
		if err != nil {
			return val
		}

		val["full_match"] = data
	}

	if len(rce.partialMatch) > 0 {
		data, err := json.Marshal(rce.partialMatch)
		if err != nil {
			return val
		}

		val["full_match"] = data
	}
	return val
}

type bootstrapEvent struct {
	bootstrap.Config
	externalID string
	success    bool
}

func (be bootstrapEvent) encode() map[string]interface{} {
	val := map[string]interface{}{
		"external_id": be.externalID,
		"success":     be.success,
		"operation":   thingBootstrap,
	}

	if be.MFThing != "" {
		val["mainflux_thing"] = be.MFThing
	}
	if be.Content != "" {
		val["content"] = be.Content
	}
	if be.Owner != "" {
		val["owner"] = be.Owner
	}
	if be.Name != "" {
		val["name"] = be.Name
	}
	if be.ExternalID != "" {
		val["external_id"] = be.ExternalID
	}
	if len(be.MFChannels) > 0 {
		channels := make([]string, len(be.MFChannels))
		for i, ch := range be.MFChannels {
			channels[i] = ch.ID
		}
		val["channels"] = fmt.Sprintf("[%s]", strings.Join(channels, ", "))
	}
	if be.ClientCert != "" {
		val["client_cert"] = be.ClientCert
	}
	if be.ClientKey != "" {
		val["client_key"] = be.ClientKey
	}
	if be.CACert != "" {
		val["ca_cert"] = be.CACert
	}
	if be.Content != "" {
		val["content"] = be.Content
	}
	return val
}

type changeStateEvent struct {
	mfThing string
	state   bootstrap.State
}

func (cse changeStateEvent) encode() map[string]interface{} {
	return map[string]interface{}{
		"thing_id":  cse.mfThing,
		"state":     cse.state.String(),
		"operation": thingStateChange,
	}
}

type updateConnectionsEvent struct {
	mfThing    string
	mfChannels []string
}

func (uce updateConnectionsEvent) encode() map[string]interface{} {
	return map[string]interface{}{
		"thing_id":  uce.mfThing,
		"channels":  fmt.Sprintf("[%s]", strings.Join(uce.mfChannels, ", ")),
		"operation": thingUpdateConnections,
	}
}

type updateCertEvent struct {
	thingKey, clientCert, clientKey, caCert string
}

func (uce updateCertEvent) encode() map[string]interface{} {
	return map[string]interface{}{
		"thing_key":   uce.thingKey,
		"client_cert": uce.clientCert,
		"client_key":  uce.clientKey,
		"ca_cert":     uce.caCert,
		"operation":   certUpdate,
	}
}

type removeHandlerEvent struct {
	id        string
	operation string
}

func (rhe removeHandlerEvent) encode() map[string]interface{} {
	return map[string]interface{}{
		"config_id": rhe.id,
		"operation": rhe.operation,
	}
}

type updateChannelHandlerEvent struct {
	bootstrap.Channel
}

func (uche updateChannelHandlerEvent) encode() map[string]interface{} {
	val := map[string]interface{}{
		"operation": channelUpdateHandler,
	}

	if uche.ID != "" {
		val["channel_id"] = uche.ID
	}
	if uche.Name != "" {
		val["name"] = uche.Name
	}
	if uche.Metadata != nil {
		metadata, err := json.Marshal(uche.Metadata)
		if err != nil {
			return val
		}

		val["metadata"] = metadata
	}
	return val
}

type disconnectThingEvent struct {
	thingID   string
	channelID string
}

func (dte disconnectThingEvent) encode() map[string]interface{} {
	return map[string]interface{}{
		"thing_id":   dte.thingID,
		"channel_id": dte.channelID,
		"operation":  thingDisconnect,
	}
}
