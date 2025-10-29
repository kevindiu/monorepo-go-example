//
// Copyright (C) 2025 Kevin Diu <kevindiujp@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package errors

import (
	"fmt"
	"runtime"
)

// Error represents a custom error with additional context
type Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Cause   error  `json:"cause,omitempty"`
	Stack   string `json:"stack,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Cause
}

// New creates a new error with stack trace
func New(message string) error {
	return &Error{
		Message: message,
		Stack:   getStack(),
	}
}

// Newf creates a new error with formatted message
func Newf(format string, args ...interface{}) error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Stack:   getStack(),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &Error{
		Message: message,
		Cause:   err,
		Stack:   getStack(),
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
		Stack:   getStack(),
	}
}

// WithCode adds an error code to an error
func WithCode(err error, code string) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(*Error); ok {
		e.Code = code
		return e
	}

	return &Error{
		Code:    code,
		Message: err.Error(),
		Stack:   getStack(),
	}
}

// GetCode extracts error code from error
func GetCode(err error) string {
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return ""
}

// Common error codes
const (
	CodeNotFound     = "NOT_FOUND"
	CodeInvalidInput = "INVALID_INPUT"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeInternal     = "INTERNAL_ERROR"
	CodeConflict     = "CONFLICT"
	CodeUnavailable  = "UNAVAILABLE"
)

// Predefined errors
var (
	ErrNotFound     = &Error{Code: CodeNotFound, Message: "resource not found"}
	ErrInvalidInput = &Error{Code: CodeInvalidInput, Message: "invalid input"}
	ErrUnauthorized = &Error{Code: CodeUnauthorized, Message: "unauthorized"}
	ErrForbidden    = &Error{Code: CodeForbidden, Message: "forbidden"}
	ErrInternal     = &Error{Code: CodeInternal, Message: "internal server error"}
	ErrConflict     = &Error{Code: CodeConflict, Message: "resource conflict"}
	ErrUnavailable  = &Error{Code: CodeUnavailable, Message: "service unavailable"}
)

func getStack() string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, 2*len(buf))
	}
}
