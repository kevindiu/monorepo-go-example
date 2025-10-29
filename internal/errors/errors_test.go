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
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	msg := "test error"
	err := New(msg)
	if err == nil {
		t.Error("New() returned nil")
	}
	if err.Error() != msg {
		t.Errorf("New() error message = %v, want %v", err.Error(), msg)
	}
}

func TestWrap(t *testing.T) {
	baseErr := errors.New("base error")
	msg := "wrapped"

	err := Wrap(baseErr, msg)
	if err == nil {
		t.Error("Wrap() returned nil")
	}

	// Should contain both messages
	errStr := err.Error()
	if errStr == "" {
		t.Error("Wrap() returned empty error message")
	}
}

func TestWrapf(t *testing.T) {
	baseErr := errors.New("base error")

	err := Wrapf(baseErr, "wrapped: %s", "formatted")
	if err == nil {
		t.Error("Wrapf() returned nil")
	}

	errStr := err.Error()
	if errStr == "" {
		t.Error("Wrapf() returned empty error message")
	}
}

func TestWithCode(t *testing.T) {
	err := New("test error")
	codedErr := WithCode(err, CodeInvalidInput)

	if codedErr == nil {
		t.Fatal("WithCode returned nil")
	}

	errObj, ok := codedErr.(*Error)
	if !ok {
		t.Fatal("WithCode did not return *Error")
	}

	if errObj.Code != CodeInvalidInput {
		t.Errorf("WithCode() code = %v, want %v", errObj.Code, CodeInvalidInput)
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "error with code",
			err:  WithCode(New("test"), CodeInvalidInput),
			want: CodeInvalidInput,
		},
		{
			name: "error without code",
			err:  New("test"),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCode(tt.err)
			if got != tt.want {
				t.Errorf("GetCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorMessage(t *testing.T) {
	err := &Error{
		Code:    CodeInvalidInput,
		Message: "invalid email",
	}

	errStr := err.Error()
	if errStr != "invalid email" {
		t.Errorf("Error.Error() = %v, want %v", errStr, "invalid email")
	}

	// Test with cause
	cause := New("original error")
	errWithCause := Wrap(cause, "wrapper message")
	errStr = errWithCause.Error()
	if errStr == "" {
		t.Error("Error.Error() returned empty string")
	}
}
