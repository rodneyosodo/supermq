// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package clients

// Emailer wrapper around the email.
type Emailer interface {
	// SendPasswordReset sends an email to the user with a link to reset the password.
	SendPasswordReset(to, host, userName, token string) error

	// SendInvitation sends an email to the user with a link to accept the invitation.
	SendInvitation(to, host, token string) error
}
