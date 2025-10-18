/*
Copyright 2025 Pextra Inc.

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

// API error struct.
type APIError struct {
	Err     error  `json:"-"` // not serialized
	Status  int    `json:"code"`
	Message string `json:"message"`
}

var _ error = (*APIError)(nil) // compile-time check

// Implements the error interface.
func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Returns the underlying error for use with errors.Is / errors.As.
func (e *APIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Creates a new APIError with status and message.
func NewAPIError(status int, message string) *APIError {
	return &APIError{
		Err:     nil,
		Status:  status,
		Message: message,
	}
}

// Wraps an existing error into an APIError with status and message.
func WrapAPIError(err error, status int, message string) *APIError {
	return &APIError{
		Err:     err,
		Status:  status,
		Message: message,
	}
}
