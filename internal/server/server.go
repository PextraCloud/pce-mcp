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
package server

import (
	"context"
	"log"
	"os"

	"github.com/PextraCloud/pce-mcp/internal/session"
	"github.com/mark3labs/mcp-go/server"
)

func GetServer() *server.MCPServer {
	hooks := &server.Hooks{}
	// Session hooks
	hooks.AddOnRegisterSession(func(ctx context.Context, s server.ClientSession) {
		sessionID := s.SessionID()
		if err := session.RegisterSession(sessionID); err != nil {
			log.Printf("failed to register session %s: %v", sessionID, err)
		}
	})
	hooks.AddOnUnregisterSession(func(ctx context.Context, s server.ClientSession) {
		session.UnregisterSession(s.SessionID())
	})

	s := server.NewMCPServer("Pextra CloudEnvironment(R) MCP Server", "1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	return s
}

func StartStdio(s *server.MCPServer) error {
	stdioServer := server.NewStdioServer(s)
	return stdioServer.Listen(context.Background(), os.Stdin, os.Stdout)
}

func StartSSE(s *server.MCPServer, addr string) error {
	sseServer := server.NewSSEServer(s)
	return sseServer.Start(addr)
}

func StartStreamableHTTP(s *server.MCPServer, addr string) error {
	httpServer := server.NewStreamableHTTPServer(s)
	return httpServer.Start(addr)
}
