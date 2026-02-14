// Copyright (c) 2025 anthoniech
// Licensed under the MIT License. See LICENSE file for details.

package mcp

import (
	"net/http"

	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	mcpServer *server.MCPServer
}

func New(pveURL, pveToken string) (*Server, error) {
	client := NewProxmoxClient(pveURL, pveToken)

	s := server.NewMCPServer("proxmox", "1.0.0",
		server.WithToolCapabilities(true),
	)

	RegisterClusterTools(s, client)
	RegisterGuestTools(s, client)
	RegisterCreateTools(s, client)
	RegisterSnapshotTools(s, client)
	RegisterBackupTools(s, client)
	RegisterStorageTools(s, client)
	RegisterTaskTools(s, client)

	return &Server{mcpServer: s}, nil
}

func (m *Server) Start() {
	log.Info("Starting MCP server (stdio)")
	if err := server.ServeStdio(m.mcpServer); err != nil {
		log.Errorf("MCP server error: %v", err)
	}
}

func (m *Server) Handler() http.Handler {
	return server.NewStreamableHTTPServer(m.mcpServer)
}

func (m *Server) MCPServer() *server.MCPServer {
	return m.mcpServer
}

func (m *Server) Close() {
	log.Info("Stopping MCP server...")
}
