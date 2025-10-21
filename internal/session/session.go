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
package session

import (
	"fmt"
	"time"

	"github.com/PextraCloud/pce-mcp/internal/config"
	"github.com/PextraCloud/pce-mcp/pkg/api"
)

// <session id> -> *api.Client
var sessionStore = make(map[string]*api.Client)

func getApiClient(id string) (*api.Client, error) {
	c := config.Get()
	timeout := c.PCEDefaultTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client, err := api.NewClient(c.PCEBaseURL, c.PCEInsecureTLS, timeout, c.PCECACertPath, c.PCECustomHeaders)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}

	return client, nil
}

func RegisterSession(id string) error {
	client, err := getApiClient(id)
	if err != nil {
		return err
	}
	sessionStore[id] = client
	return nil
}

func UnregisterSession(id string) {
	delete(sessionStore, id)
}

func GetSession(id string) (*api.Client, error) {
	return getApiClient(id)
	// TODO implement session management
	// client, exists := sessionStore["id"]
	//
	//	if !exists {
	//		return nil, fmt.Errorf("session not found")
	//	}
	//
	// return client, nil
}
