// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import (
	"errors"
	"fmt"
)

// Error is returned for any failed API call (non-2xx HTTP, or status="fail").
type Error struct {
	HTTPStatus int    `json:"-"`
	Code       int    `json:"code"`
	Status     string `json:"status,omitempty"`
	Message    string `json:"message,omitempty"`
}

func (e *Error) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("loadmaster: HTTP %d (code=%d status=%s)", e.HTTPStatus, e.Code, e.Status)
	}
	return fmt.Sprintf("loadmaster: %s (HTTP %d code=%d)", e.Message, e.HTTPStatus, e.Code)
}

// IsNotFound reports whether err was caused by an "Unknown VS"-style
// response from LoadMaster, used by Read handlers to drop deleted resources
// from state cleanly.
func IsNotFound(err error) bool {
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.Message {
	case "Unknown VS", "Unknown RS", "Unknown SubVS", "Unknown ACME cert":
		return true
	}
	return false
}
