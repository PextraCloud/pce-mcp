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

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	HTTP      *http.Client
	BaseURL   *url.URL
	APIPrefix string
	Headers   http.Header
}

// Construct a new Client.
func NewClient(baseURL string, insecureSkipVerify bool, timeout time.Duration, caCertPath string, customHeaders http.Header) (*Client, error) {
	if insecureSkipVerify && caCertPath != "" {
		return nil, WrapAPIError(fmt.Errorf("insecure skip verify and custom CA are mutually exclusive"), 0, "invalid TLS configuration")
	}

	u, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return nil, WrapAPIError(err, 0, "invalid base URL")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}

	// If CA cert path was provided, load and append to system cert pool.
	if caCertPath != "" {
		certPEM, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, WrapAPIError(err, 0, "failed to read CA cert file")
		}
		pool, err := x509.SystemCertPool()
		if err != nil {
			// If system pool isn't available, create a new one.
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(certPEM) {
			return nil, WrapAPIError(fmt.Errorf("failed to parse CA cert file"), 0, "invalid CA certificate")
		}
		tlsConfig.RootCAs = pool
		tlsConfig.InsecureSkipVerify = false
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpClient := &http.Client{Transport: transport, Timeout: timeout}

	// Custom HTTP headers
	headers := make(http.Header)
	for k, vals := range customHeaders {
		for _, v := range vals {
			headers.Add(k, v)
		}
	}
	return &Client{
		HTTP:      httpClient,
		BaseURL:   u,
		APIPrefix: "/api",
		Headers:   headers,
	}, nil
}

// newURL builds the full URL for the request path (path may be "images" or "nodes/123").
func (c *Client) newURL(path string, query url.Values) string {
	p := strings.TrimLeft(path, "/")
	base := strings.TrimRight(c.BaseURL.String(), "/")
	full := base + strings.TrimRight(c.APIPrefix, "/") + "/" + p
	if len(query) > 0 {
		full = full + "?" + query.Encode()
	}
	return full
}

// ExpandPath replaces placeholders in the form {name} with the corresponding
// values from params. Values are url.PathEscaped. Example:
//
//	p := client.ExpandPath("/nodes/{node_id}/images", map[string]string{"node_id": id})
//	client.Get(ctx, p, nil, out)
func (c *Client) ExpandPath(pathTemplate string, params map[string]string) string {
	for k, v := range params {
		placeholder := "{" + k + "}"
		pathTemplate = strings.ReplaceAll(pathTemplate, placeholder, url.PathEscape(v))
	}
	return pathTemplate
}

// newRequest creates an http.Request and applies common headers.
func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values, body io.Reader) (*http.Request, *APIError) {
	full := c.newURL(path, query)
	req, err := http.NewRequestWithContext(ctx, method, full, body)
	if err != nil {
		return nil, WrapAPIError(err, 0, "creating request")
	}
	req.Header.Set("Accept", "application/json")
	for k, vv := range c.Headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	// Authentication headers can be provided via c.Headers by callers.
	// Avoid debug prints in production; use structured logging at call sites if needed.

	// caller may set Content-Type when body is provided
	return req, nil
}

// Do executes the request and decodes JSON into out if provided.
// On non-2xx responses it tries to parse an APIError from the body.
func (c *Client) Do(req *http.Request, out any) *APIError {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return WrapAPIError(err, 0, "request failed")
	}
	defer resp.Body.Close()

	// Handle error status codes.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		// try decode APIError from body
		var parsed APIError
		if len(data) > 0 && json.Unmarshal(data, &parsed) == nil && (parsed.Status != 0 || parsed.Message != "") {
			// ensure Status set
			if parsed.Status == 0 {
				parsed.Status = resp.StatusCode
			}
			return &parsed
		}
		// fallback: raw body -> message
		msg := strings.TrimSpace(string(data))
		if msg == "" {
			msg = resp.Status
		}
		return &APIError{Err: fmt.Errorf("%s", msg), Status: resp.StatusCode, Message: msg}
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return WrapAPIError(err, resp.StatusCode, "decoding response")
		}
	}
	return nil
}

// Get convenience helper to perform a GET and decode JSON response into out.
func (c *Client) Get(ctx context.Context, path string, query url.Values, out any) *APIError {
	req, apiErr := c.newRequest(ctx, http.MethodGet, path, query, nil)
	if apiErr != nil {
		return apiErr
	}
	return c.Do(req, out)
}

func (c *Client) Post(ctx context.Context, path string, query url.Values, body io.Reader, out any) *APIError {
	req, apiErr := c.newRequest(ctx, http.MethodPost, path, query, body)
	if apiErr != nil {
		return apiErr
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.Do(req, out)
}

func (c *Client) Put(ctx context.Context, path string, query url.Values, body io.Reader, out any) *APIError {
	req, apiErr := c.newRequest(ctx, http.MethodPut, path, query, body)
	if apiErr != nil {
		return apiErr
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.Do(req, out)
}

func (c *Client) Delete(ctx context.Context, path string, query url.Values, out any) *APIError {
	req, apiErr := c.newRequest(ctx, http.MethodDelete, path, query, nil)
	if apiErr != nil {
		return apiErr
	}
	return c.Do(req, out)
}
