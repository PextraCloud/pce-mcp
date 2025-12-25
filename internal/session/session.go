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
	"sync"
	"time"

	"github.com/PextraCloud/pce-mcp/internal/config"
	"github.com/PextraCloud/pce-mcp/pkg/api"
)

type sessionEntry struct {
	client *api.Client
	mu     sync.Mutex
}

// <session id> -> *api.Client
var (
	sessionStore = make(map[string]*sessionEntry)
	sessionMu    sync.RWMutex
)

func getApiClient() (*api.Client, error) {
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
	if id == "" {
		return fmt.Errorf("session id is required")
	}
	entry, err := newSessionEntry()
	if err != nil {
		return err
	}
	sessionMu.Lock()
	sessionStore[id] = entry
	sessionMu.Unlock()
	return nil
}

func UnregisterSession(id string) {
	sessionMu.Lock()
	delete(sessionStore, id)
	sessionMu.Unlock()
}

func GetSession(id string, authorization string) (*api.Client, error) {
	if id == "" {
		return nil, fmt.Errorf("session id is required")
	}

	sessionMu.RLock()
	entry, ok := sessionStore[id]
	sessionMu.RUnlock()
	if !ok {
		var err error
		entry, err = newSessionEntry()
		if err != nil {
			return nil, err
		}
		sessionMu.Lock()
		sessionStore[id] = entry
		sessionMu.Unlock()
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()
	setAuthorization(entry.client, authorization)
	return entry.client, nil
}

func newSessionEntry() (*sessionEntry, error) {
	client, err := getApiClient()
	if err != nil {
		return nil, err
	}
	return &sessionEntry{client: client}, nil
}

func setAuthorization(client *api.Client, authorization string) {
	if client == nil {
		return
	}
	if authorization == "" {
		client.Headers.Del("Authorization")
		return
	}
	client.Headers.Set("Authorization", authorization)
}
