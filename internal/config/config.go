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
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	EnvSSEAddr       = "SSE_ADDR"
	EnvHTTPAddr      = "HTTP_ADDR"
	EnvBaseURL       = "BASE_URL"
	EnvTLSSkipVerify = "TLS_SKIP_VERIFY"
	EnvTimeout       = "TIMEOUT"
)

// AppConfig holds runtime configuration for the server and API client.
type AppConfig struct {
	// Listen addresses
	SSEAddr  string
	HTTPAddr string

	// PCE API client
	PCEBaseURL        string
	PCEInsecureTLS    bool
	PCEDefaultTimeout time.Duration
}

var cfg AppConfig

// Set sets the global configuration; typically called by CLI before server start.
func Set(c AppConfig) { cfg = c }

// Get returns the current configuration.
func Get() AppConfig { return cfg }

// WithEnvDefaults returns a copy of c with environment variable fallbacks applied
// and performs validation. Returns an error if validation fails.
func WithEnvDefaults(c AppConfig) (*AppConfig, error) {
	// apply env fallbacks
	if c.SSEAddr == "" {
		if v := os.Getenv(EnvSSEAddr); v != "" {
			c.SSEAddr = v
		}
	}
	if c.HTTPAddr == "" {
		if v := os.Getenv(EnvHTTPAddr); v != "" {
			c.HTTPAddr = v
		}
	}
	if c.PCEBaseURL == "" {
		if v := os.Getenv(EnvBaseURL); v != "" {
			c.PCEBaseURL = v
		}
	}

	// Booleans: apply env if provided (validate on parse failure)
	if v := os.Getenv(EnvTLSSkipVerify); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.PCEInsecureTLS = b
		} else {
			return nil, validationError{msgs: []string{fmt.Sprintf("invalid %s: %s", EnvTLSSkipVerify, v)}}
		}
	}
	// Timeout: env override if provided (validate on parse failure)
	if c.PCEDefaultTimeout <= 0 {
		if v := os.Getenv(EnvTimeout); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				c.PCEDefaultTimeout = time.Duration(n) * time.Second
			} else {
				return nil, validationError{msgs: []string{fmt.Sprintf("invalid %s: %s", EnvTimeout, v)}}
			}
		}
	}

	// collect validation issues
	errs := []string{}

	// Require at least one listen address
	if c.HTTPAddr == "" && c.SSEAddr == "" {
		errs = append(errs, "at least one of %s or %s must be set", EnvHTTPAddr, EnvSSEAddr)
	}

	// PCE base URL required and must look like http:// or https://
	if c.PCEBaseURL == "" {
		errs = append(errs, fmt.Sprintf("%s is required", EnvBaseURL))
	} else {
		_, err := url.ParseRequestURI(c.PCEBaseURL)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s is invalid: %v", EnvBaseURL, err))
		}
	}

	// Timeout must be positive
	if c.PCEDefaultTimeout <= 0 {
		errs = append(errs, "%s must be > 0 (seconds)", EnvTimeout)
	}

	if len(errs) > 0 {
		return nil, validationError{msgs: errs}
	}
	return &c, nil
}

// validationError collects validation messages.
type validationError struct {
	msgs []string
}

func (e validationError) Error() string {
	if len(e.msgs) == 0 {
		return "validation failed"
	}
	out := e.msgs[0]
	for i := 1; i < len(e.msgs); i++ {
		out += "; " + e.msgs[i]
	}
	return out
}
