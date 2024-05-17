// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"net/http"
)

const (
	// LevelInfo represents an informational message.
	LevelInfo uint8 = iota
	// LevelWarn represents a warning message.
	LevelWarn
	// LevelError represents an error message.
	LevelError
	// LevelFatal represents a fatal message.
	LevelFatal
)

// Error specifies an API that must be fullfiled by error type.
type Error interface {
	// Error implements the error interface.
	Error() string

	// Msg returns error message.
	Msg() string

	// Err returns wrapped error.
	Err() Error

	// MarshalJSON returns a marshaled error.
	MarshalJSON() ([]byte, error)

	// Level returns the level of the error.
	// If two or more errors are wrapped, the highest level is returned.
	Level() uint8

	// HTTPStatusCode returns the HTTP status code of the error.
	// If two or more errors are wrapped, the highest status code is returned.
	HTTPStatusCode() uint16
}

var _ Error = (*customError)(nil)

// customError represents a Magistrala error.
type customError struct {
	msg        string
	level      uint8
	statusCode uint16
	err        Error
}

type Option func(*customError)

func WithLevel(level uint8) Option {
	return func(ce *customError) {
		ce.level = level
	}
}

func WithCode(code uint16) Option {
	return func(ce *customError) {
		ce.statusCode = code
	}
}

// New returns an Error that formats as the given text.
func New(text string, opts ...Option) Error {
	ce := &customError{
		msg:        text,
		err:        nil,
		level:      LevelInfo,
		statusCode: http.StatusBadRequest,
	}

	for _, opt := range opts {
		opt(ce)
	}

	return ce
}

func (ce *customError) Error() string {
	if ce == nil {
		return ""
	}
	if ce.err == nil {
		return ce.msg
	}
	return ce.msg + " : " + ce.err.Error()
}

func (ce *customError) Msg() string {
	return ce.msg
}

func (ce *customError) Err() Error {
	return ce.err
}

func (ce *customError) MarshalJSON() ([]byte, error) {
	var val string
	if e := ce.Err(); e != nil {
		val = e.Msg()
	}
	return json.Marshal(&struct {
		Err  string `json:"error"`
		Msg  string `json:"message"`
		Lvl  uint8  `json:"level"`
		Code uint16 `json:"status_code"`
	}{
		Err:  val,
		Msg:  ce.Msg(),
		Lvl:  ce.Level(),
		Code: ce.HTTPStatusCode(),
	})
}

func (ce *customError) Level() uint8 {
	if ce.err == nil {
		return ce.level
	}

	if ce.level > ce.err.Level() {
		return ce.level
	}

	return ce.err.Level()
}

func (ce *customError) HTTPStatusCode() uint16 {
	if ce.err == nil {
		return ce.statusCode
	}

	if ce.statusCode > ce.err.HTTPStatusCode() {
		return ce.statusCode
	}

	return ce.err.HTTPStatusCode()
}

// Contains inspects if e2 error is contained in any layer of e1 error.
func Contains(e1, e2 error) bool {
	if e1 == nil || e2 == nil {
		return e2 == e1
	}
	ce, ok := e1.(Error)
	if ok {
		if ce.Msg() == e2.Error() {
			return true
		}
		return Contains(ce.Err(), e2)
	}
	return e1.Error() == e2.Error()
}

// Wrap returns an Error that wrap err with wrapper.
func Wrap(wrapper, err error) Error {
	if wrapper == nil || err == nil {
		return New(wrapper.Error())
	}
	if w, ok := wrapper.(Error); ok {
		return &customError{
			msg:        w.Msg(),
			err:        cast(err),
			level:      getLevel(w, cast(err)),
			statusCode: getStatusCode(w, cast(err)),
		}
	}
	return &customError{
		msg:        wrapper.Error(),
		err:        cast(err),
		level:      getLevel(wrapper, cast(err)),
		statusCode: getStatusCode(wrapper, cast(err)),
	}
}

// Unwrap returns the wrapper and the error by separating the Wrapper from the error.
func Unwrap(err error) (error, error) {
	if ce, ok := err.(Error); ok {
		if ce.Err() == nil {
			return nil, New(ce.Msg())
		}
		return New(ce.Msg()), ce.Err()
	}

	return nil, err
}

func cast(err error) Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(Error); ok {
		return e
	}
	return &customError{
		msg: err.Error(),
		err: nil,
	}
}

func getLevel(err ...error) uint8 {
	var lvl uint8
	for _, e := range err {
		if e == nil {
			continue
		}
		if ce, ok := e.(Error); ok {
			if ce.Level() > lvl {
				lvl = ce.Level()
			}
		}
	}

	return lvl
}

func getStatusCode(err ...error) uint16 {
	var code uint16
	for _, e := range err {
		if e == nil {
			continue
		}
		if ce, ok := e.(Error); ok {
			if ce.HTTPStatusCode() > code {
				code = ce.HTTPStatusCode()
			}
		}
	}

	return code
}
