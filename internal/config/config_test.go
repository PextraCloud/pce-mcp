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
package config

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestReadEnvString(t *testing.T) {
	envVarName := "TEST_ENV_STRING"
	fallbackValue := "_FallbackValue"

	tests := []struct {
		name     string
		expected string
	}{
		{"set", "127.0.0.1:8080"},
		{"empty", fallbackValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(envVarName, tt.expected)
			defer os.Unsetenv(envVarName)

			got := readEnvString(envVarName, fallbackValue)
			if got != tt.expected {
				t.Errorf("readEnvString(%s) = %s, want %s", envVarName, got, tt.expected)
			}
		})
	}
}

func TestWithEnvDefaults(t *testing.T) {
	tests := []struct {
		name     string
		initial  AppConfig
		env      map[string]string
		expected *AppConfig
		err      *validationError
	}{
		{
			name: "Valid HTTP_ADDR only",
			initial: AppConfig{
				PCEBaseURL:        "https://api.example.com",
				PCEDefaultTimeout: 30 * time.Second,
			},
			env: map[string]string{
				EnvHTTPAddr: ":8080",
			},
			expected: &AppConfig{
				HTTPAddr:          ":8080",
				PCEBaseURL:        "https://api.example.com",
				PCEDefaultTimeout: 30 * time.Second,
			},
			err: nil,
		},
		{
			name: "Disable stdio false with no listen addresses succeeds",
			initial: AppConfig{
				PCEBaseURL:        "https://api.example.com",
				PCEDefaultTimeout: 5 * time.Second,
			},
			env: map[string]string{
				EnvDisableStdio: "false",
			},
			expected: &AppConfig{
				PCEBaseURL:        "https://api.example.com",
				PCEDefaultTimeout: 5 * time.Second,
				DisableStdio:      false,
			},
			err: nil,
		},
		{
			name: "Disable stdio true with no listeners fails",
			initial: AppConfig{
				PCEBaseURL:        "https://api.example.com",
				PCEDefaultTimeout: 10 * time.Second,
			},
			env: map[string]string{
				EnvDisableStdio: "true",
			},
			expected: nil,
			err: &validationError{
				msgs: []string{
					fmt.Sprintf("at least one of %s, %s, or stdio server must be enabled", EnvSSEAddr, EnvHTTPAddr),
				},
			},
		},
		{
			name: "Missing base URL",
			initial: AppConfig{
				PCEDefaultTimeout: 10 * time.Second,
			},
			env:      map[string]string{},
			expected: nil,
			err: &validationError{
				msgs: []string{
					fmt.Sprintf("%s is required", EnvBaseURL),
				},
			},
		},
		{
			name: "Invalid base URL",
			initial: AppConfig{
				PCEBaseURL:        "://bad-url",
				PCEDefaultTimeout: 10 * time.Second,
			},
			env:      map[string]string{},
			expected: nil,
			err: &validationError{
				msgs: []string{
					fmt.Sprintf("%s is invalid: parse \"%s\": missing protocol scheme", EnvBaseURL, "://bad-url"),
				},
			},
		},
		{
			name:    "Invalid TLS skip verify boolean",
			initial: AppConfig{},
			env: map[string]string{
				EnvTLSSkipVerify: "notabool",
			},
			expected: nil,
			err: &validationError{
				msgs: []string{
					fmt.Sprintf("invalid %s: %s", EnvTLSSkipVerify, "notabool"),
				},
			},
		},
		{
			name:    "Invalid disable stdio boolean",
			initial: AppConfig{},
			env: map[string]string{
				EnvDisableStdio: "abc",
			},
			expected: nil,
			err: &validationError{
				msgs: []string{
					fmt.Sprintf("invalid %s: %s", EnvDisableStdio, "abc"),
				},
			},
		},
		{
			name: "Timeout from environment",
			initial: AppConfig{
				PCEBaseURL: "https://api.example.com",
			},
			env: map[string]string{
				EnvTimeout: "45",
			},
			expected: &AppConfig{
				PCEBaseURL:        "https://api.example.com",
				PCEDefaultTimeout: 45 * time.Second,
			},
			err: nil,
		},
		{
			name: "Existing timeout is not overridden by environment",
			initial: AppConfig{
				PCEBaseURL:        "https://api.example.com",
				PCEDefaultTimeout: 15 * time.Second,
			},
			env: map[string]string{
				EnvTimeout: "60",
			},
			expected: &AppConfig{
				PCEBaseURL:        "https://api.example.com",
				PCEDefaultTimeout: 15 * time.Second,
			},
			err: nil,
		},
		{
			name: "Invalid timeout non numeric",
			initial: AppConfig{
				PCEBaseURL: "https://api.example.com",
			},
			env: map[string]string{
				EnvTimeout: "abc",
			},
			expected: nil,
			err: &validationError{
				msgs: []string{
					fmt.Sprintf("invalid %s: %s", EnvTimeout, "abc"),
				},
			},
		},
		{
			name: "Invalid timeout zero",
			initial: AppConfig{
				PCEBaseURL: "https://api.example.com",
			},
			env: map[string]string{
				EnvTimeout: "0",
			},
			expected: nil,
			err: &validationError{
				msgs: []string{
					fmt.Sprintf("invalid %s: %s", EnvTimeout, "0"),
				},
			},
		},
		{
			name: "Invalid timeout negative",
			initial: AppConfig{
				PCEBaseURL: "https://api.example.com",
			},
			env: map[string]string{
				EnvTimeout: "-5",
			},
			expected: nil,
			err: &validationError{
				msgs: []string{
					fmt.Sprintf("invalid %s: %s", EnvTimeout, "-5"),
				},
			},
		},
		{
			name: "Missing timeout after validation",
			initial: AppConfig{
				PCEBaseURL: "https://api.example.com",
			},
			env:      map[string]string{},
			expected: nil,
			err: &validationError{
				msgs: []string{
					fmt.Sprintf("%s must be > 0 (seconds)", EnvTimeout),
				},
			},
		},
		{
			name: "TLS skip verify enabled",
			initial: AppConfig{
				PCEBaseURL:        "https://api.example.com",
				PCEDefaultTimeout: 10 * time.Second,
			},
			env: map[string]string{
				EnvTLSSkipVerify: "true",
			},
			expected: &AppConfig{
				PCEBaseURL:        "https://api.example.com",
				PCEInsecureTLS:    true,
				PCEDefaultTimeout: 10 * time.Second,
			},
			err: nil,
		},
		{
			name: "TLS skip verify and CA cert both configured",
			initial: AppConfig{
				PCEBaseURL:        "https://api.example.com",
				PCEDefaultTimeout: 10 * time.Second,
				PCECACertPath:     "/path/to/ca.pem",
			},
			env: map[string]string{
				EnvTLSSkipVerify: "true",
			},
			expected: nil,
			err: &validationError{
				msgs: []string{
					fmt.Sprintf("only one of %s or %s may be set", EnvTLSSkipVerify, EnvCACert),
					// The CA cert path does not exist, so we expect an error for that as well
					fmt.Sprintf("%s points to invalid path: stat %s: no such file or directory", EnvCACert, "/path/to/ca.pem"),
				},
			},
		},
		{
			name: "CA certificate path does not exist",
			initial: AppConfig{
				PCEBaseURL:        "https://api.example.com",
				PCECACertPath:     "/path/does/not/exist.pem",
				PCEDefaultTimeout: 10 * time.Second,
			},
			env:      map[string]string{},
			expected: nil,
			err: &validationError{
				msgs: []string{
					fmt.Sprintf("%s points to invalid path: stat %s: no such file or directory", EnvCACert, "/path/does/not/exist.pem"),
				},
			},
		},
		{
			name: "Multiple validation errors collected",
			initial: AppConfig{
				DisableStdio:      true,
				PCEBaseURL:        "",
				PCEDefaultTimeout: 0,
			},
			env:      map[string]string{},
			expected: nil,
			err: &validationError{
				msgs: []string{
					fmt.Sprintf("at least one of %s, %s, or stdio server must be enabled", EnvSSEAddr, EnvHTTPAddr),
					fmt.Sprintf("%s is required", EnvBaseURL),
					fmt.Sprintf("%s must be > 0 (seconds)", EnvTimeout),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for the test
			for key, value := range tt.env {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			recv, err := WithEnvDefaults(tt.initial)
			if err != nil {
				if tt.err == nil {
					t.Errorf("WithEnvDefaults() unexpected error = %v", err)
				} else if err.Error() != tt.err.Error() {
					t.Errorf("WithEnvDefaults() error = %v, want %v", err, tt.err)
				}
				return
			} else if tt.err != nil {
				t.Errorf("WithEnvDefaults() expected error = %v, got nil", tt.err)
				return
			}

			if !reflect.DeepEqual(*recv, *tt.expected) {
				t.Errorf("WithEnvDefaults() got = %v, want %v", *recv, *tt.expected)
			}
		})
	}
}

func TestGetAndSetConfig(t *testing.T) {
	testCfg := AppConfig{
		SSEAddr:           "127.0.0.1:8080",
		PCEBaseURL:        "https://api.example.com",
		PCEDefaultTimeout: 30 * time.Second,
	}

	Set(testCfg)
	got := Get()

	if !reflect.DeepEqual(got, testCfg) {
		t.Errorf("Get() = %v, want %v", got, testCfg)
	}
}
