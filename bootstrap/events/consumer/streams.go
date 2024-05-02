// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package consumer

import (
	"context"
	"time"

	"github.com/absmach/magistrala/bootstrap"
	"github.com/absmach/magistrala/pkg/events"
)

const (
	thingRemove = "thing.remove"

	channelPrefix   = "channel."
	channelUpdate   = channelPrefix + "update"
	channelRemove   = channelPrefix + "remove"
	thingDisconnect = channelPrefix + "unassign"
)

type eventHandler struct {
	svc bootstrap.Service
}

// NewEventHandler returns new event store handler.
func NewEventHandler(svc bootstrap.Service) events.EventHandler {
	return &eventHandler{
		svc: svc,
	}
}

func (es *eventHandler) Handle(ctx context.Context, event events.Event) error {
	msg, err := event.Encode()
	if err != nil {
		return err
	}

	switch msg["operation"] {
	case thingRemove:
		rte := decodeRemoveThing(msg)
		err = es.svc.RemoveConfigHandler(ctx, rte.id)
	case thingDisconnect:
		dte := decodeDisconnectThing(msg)

		for _, thingID := range dte.thingIDs {
			if err = es.svc.DisconnectThingHandler(ctx, dte.channelID, thingID); err != nil {
				return err
			}
		}
	case channelUpdate:
		uce := decodeUpdateChannel(msg)
		err = es.handleUpdateChannel(ctx, uce)
	case channelRemove:
		rce := decodeRemoveChannel(msg)
		err = es.svc.RemoveChannelHandler(ctx, rce.id)
	}
	if err != nil {
		return err
	}

	return nil
}

func decodeRemoveThing(event map[string]interface{}) removeEvent {
	return removeEvent{
		id: events.Read(event, "id", ""),
	}
}

func decodeUpdateChannel(event map[string]interface{}) updateChannelEvent {
	metadata := events.Read(event, "metadata", map[string]interface{}{})

	return updateChannelEvent{
		id:        events.Read(event, "id", ""),
		name:      events.Read(event, "name", ""),
		metadata:  metadata,
		updatedAt: events.Read(event, "updated_at", time.Now()),
		updatedBy: events.Read(event, "updated_by", ""),
	}
}

func decodeRemoveChannel(event map[string]interface{}) removeEvent {
	return removeEvent{
		id: events.Read(event, "id", ""),
	}
}

func decodeDisconnectThing(event map[string]interface{}) disconnectEvent {
	return disconnectEvent{
		channelID: events.Read(event, "group_id", ""),
		thingIDs:  events.ReadStringSlice(event, "member_ids"),
	}
}

func (es *eventHandler) handleUpdateChannel(ctx context.Context, uce updateChannelEvent) error {
	channel := bootstrap.Channel{
		ID:        uce.id,
		Name:      uce.name,
		Metadata:  uce.metadata,
		UpdatedAt: uce.updatedAt,
		UpdatedBy: uce.updatedBy,
	}
	return es.svc.UpdateChannelHandler(ctx, channel)
}
