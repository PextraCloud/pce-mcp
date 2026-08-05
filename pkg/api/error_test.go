/*
Copyright 2026 Pextra Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package api

import (
	"errors"
	"reflect"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "<nil>",
		},
		{
			name: "with underlying error",
			err: &APIError{
				Err:     errors.New("underlying error"),
				Status:  500,
				Message: "message",
			},
			expected: "underlying error",
		},
		{
			name: "no underlying error but has message",
			err: &APIError{
				Err:     nil,
				Status:  404,
				Message: "not found",
			},
			expected: "not found",
		},
		{
			name: "no underlying error and empty message",
			err: &APIError{
				Err:     nil,
				Status:  500,
				Message: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("APIError.Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected error
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: nil,
		},
		{
			name: "with underlying error",
			err: &APIError{
				Err:     errors.New("underlying error"),
				Status:  500,
				Message: "message",
			},
			expected: errors.New("underlying error"),
		},
		{
			name: "no underlying error",
			err: &APIError{
				Err:     nil,
				Status:  404,
				Message: "not found",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Unwrap()
			if result == nil {
				if tt.expected != nil {
					t.Errorf("APIError.Unwrap() = %v, want %v", result, tt.expected)
				}
			} else {
				if tt.expected == nil {
					t.Errorf("APIError.Unwrap() = %v, want %v", result, tt.expected)
				} else if result.Error() != tt.expected.Error() {
					t.Errorf("APIError.Unwrap() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestNewAPIError(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		message  string
		expected *APIError
	}{
		{
			name:    "standard error",
			status:  500,
			message: "internal server error",
			expected: &APIError{
				Err:     nil,
				Status:  500,
				Message: "internal server error",
			},
		},
		{
			name:    "not found error",
			status:  404,
			message: "resource not found",
			expected: &APIError{
				Err:     nil,
				Status:  404,
				Message: "resource not found",
			},
		},
		{
			name:    "bad request",
			status:  400,
			message: "invalid input",
			expected: &APIError{
				Err:     nil,
				Status:  400,
				Message: "invalid input",
			},
		},
		{
			name:    "empty message",
			status:  503,
			message: "",
			expected: &APIError{
				Err:     nil,
				Status:  503,
				Message: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewAPIError(tt.status, tt.message)
			// Okay to use reflect.DeepEqual with errors since we are effectively comparing structs
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("NewAPIError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWrapAPIError(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name     string
		err      error
		status   int
		message  string
		expected *APIError
	}{
		{
			name:    "wrap error with status",
			err:     originalErr,
			status:  500,
			message: "internal error",
			expected: &APIError{
				Err:     originalErr,
				Status:  500,
				Message: "internal error",
			},
		},
		{
			name:    "wrap error with different status",
			err:     originalErr,
			status:  400,
			message: "bad request",
			expected: &APIError{
				Err:     originalErr,
				Status:  400,
				Message: "bad request",
			},
		},
		{
			name:    "wrap nil error",
			err:     nil,
			status:  404,
			message: "not found",
			expected: &APIError{
				Err:     nil,
				Status:  404,
				Message: "not found",
			},
		},
		{
			name:    "wrap error with empty message",
			err:     originalErr,
			status:  503,
			message: "",
			expected: &APIError{
				Err:     originalErr,
				Status:  503,
				Message: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapAPIError(tt.err, tt.status, tt.message)
			// Okay to use reflect.DeepEqual with errors since we are effectively comparing structs
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("WrapAPIError() = %v, want %v", result, tt.expected)
			}
		})
	}
}
