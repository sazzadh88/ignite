// Package gate provides authorization functionality for Ignite.
// It implements Laravel-inspired Gates and Policies for fine-grained access control.
package gate

import "net/http"

// Response represents the result of an authorization check.
// It contains whether access is allowed, an optional message, and a status code.
type Response struct {
	Allowed bool
	Message string
	Code    int
}

// Allow creates a Response that grants access.
// An optional message can be provided to explain why access was granted.
func Allow(message ...string) *Response {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return &Response{
		Allowed: true,
		Message: msg,
		Code:    http.StatusOK,
	}
}

// Deny creates a Response that denies access with a 403 Forbidden status.
// An optional message can be provided to explain why access was denied.
func Deny(message ...string) *Response {
	msg := "This action is unauthorized."
	if len(message) > 0 {
		msg = message[0]
	}
	return &Response{
		Allowed: false,
		Message: msg,
		Code:    http.StatusForbidden,
	}
}

// DenyWithStatus creates a Response that denies access with a custom status code.
// An optional message can be provided to explain why access was denied.
func DenyWithStatus(status int, message ...string) *Response {
	msg := "This action is unauthorized."
	if len(message) > 0 {
		msg = message[0]
	}
	return &Response{
		Allowed: false,
		Message: msg,
		Code:    status,
	}
}

// DenyAsNotFound creates a Response that denies access with a 404 Not Found status.
// This is useful for hiding the existence of resources from unauthorized users.
func DenyAsNotFound(message ...string) *Response {
	msg := "Not found."
	if len(message) > 0 {
		msg = message[0]
	}
	return &Response{
		Allowed: false,
		Message: msg,
		Code:    http.StatusNotFound,
	}
}
