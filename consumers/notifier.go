// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package consumers

import (
	"errors"

	"github.com/absmach/supermq/pkg/messaging"
)

// ErrNotify wraps sending notification errors.
var ErrNotify = errors.New("error sending notification")

// Notifier represents an API for sending notification.
type Notifier interface {
	// Notify method is used to send notification for the
	// received message to the provided list of receivers.
	Notify(from string, to []string, msg *messaging.Message) error
}
