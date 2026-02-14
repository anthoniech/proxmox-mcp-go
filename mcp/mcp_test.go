package mcp_test

import (
	"context"
	"encoding/json"
	"sort"
	"testing"

	mcplib "github.com/anthoniech/proxmox-mcp-go/mcp"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	fakeURL   = "https://pve.example.com:8006"
	fakeToken = "user@pam!token=fake-secret"
)

// expectedTools lists every tool name that must be registered.
var expectedTools = []string{
	// cluster
	"get_cluster_status", "list_nodes", "get_node_status",
	// guest
	"list_vms", "list_containers", "list_cluster_resources",
	"start_guest", "stop_guest", "get_next_id",
	// create
	"create_vm", "create_container", "clone_guest",
	"delete_guest", "convert_to_template",
	// snapshot
	"list_snapshots", "create_snapshot", "rollback_snapshot", "delete_snapshot",
	// backup
	"backup_guest", "list_backups", "restore_backup",
	// storage
	"list_storage", "list_templates", "list_isos", "download_template",
	// task
	"list_tasks", "get_task_status", "get_task_log",
}

func newTestServer(t *testing.T) *mcplib.Server {
	t.Helper()

	s, err := mcplib.New(fakeURL, fakeToken)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if s == nil {
		t.Fatal("New() returned nil server")
	}

	return s
}

func TestNew(t *testing.T) {
	s := newTestServer(t)

	if s.MCPServer() == nil {
		t.Fatal("MCPServer() returned nil")
	}
}

func TestHandler(t *testing.T) {
	s := newTestServer(t)
	h := s.Handler()

	if h == nil {
		t.Fatal("Handler() returned nil")
	}
}

func TestToolsRegistered(t *testing.T) {
	s := newTestServer(t)
	tools := s.MCPServer().ListTools()

	if len(tools) != len(expectedTools) {
		t.Fatalf("expected %d tools, got %d", len(expectedTools), len(tools))
	}

	for _, name := range expectedTools {
		if _, ok := tools[name]; !ok {
			t.Errorf("tool %q not registered", name)
		}
	}
}

func TestToolsList(t *testing.T) {
	s := newTestServer(t)

	ctx := context.Background()
	session := server.NewInProcessSession(
		server.GenerateInProcessSessionID(), nil,
	)
	ctx = s.MCPServer().WithContext(ctx, session)

	msg := json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`)
	resp := s.MCPServer().HandleMessage(ctx, msg)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var rpcResp struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &rpcResp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if rpcResp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %q", rpcResp.JSONRPC)
	}

	if rpcResp.ID != 1 {
		t.Errorf("expected id 1, got %d", rpcResp.ID)
	}

	gotNames := make([]string, 0, len(rpcResp.Result.Tools))
	for _, tool := range rpcResp.Result.Tools {
		gotNames = append(gotNames, tool.Name)
	}

	sort.Strings(gotNames)

	want := make([]string, len(expectedTools))
	copy(want, expectedTools)
	sort.Strings(want)

	if len(gotNames) != len(want) {
		t.Fatalf("expected %d tools, got %d", len(want), len(gotNames))
	}

	for i := range want {
		if gotNames[i] != want[i] {
			t.Errorf("tool[%d]: expected %q, got %q", i, want[i], gotNames[i])
		}
	}
}

func TestCallTool(t *testing.T) {
	s := newTestServer(t)

	ctx := context.Background()
	session := server.NewInProcessSession(
		server.GenerateInProcessSessionID(), nil,
	)
	ctx = s.MCPServer().WithContext(ctx, session)

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "get_cluster_status",
			"arguments": map[string]any{},
		},
	}

	msg, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	resp := s.MCPServer().HandleMessage(ctx, msg)
	if resp == nil {
		t.Fatal("HandleMessage returned nil")
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	// The response should be valid JSON-RPC (either success or error result).
	// Since there's no real Proxmox backend, we expect an error in the tool
	// result, but the JSON-RPC envelope must be well-formed.
	var envelope struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("response is not valid JSON-RPC: %v", err)
	}

	if envelope.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %q", envelope.JSONRPC)
	}

	if envelope.ID != 2 {
		t.Errorf("expected id 2, got %d", envelope.ID)
	}

	// Verify the response is a JSONRPCResponse (not a protocol-level error).
	// The tool handler should return a CallToolResult with isError=true,
	// wrapped in a valid JSONRPCResponse.
	switch resp.(type) {
	case mcp.JSONRPCResponse:
		// Expected: tool executed, returned error in result content
	case *mcp.JSONRPCResponse:
		// Expected (pointer variant)
	case mcp.JSONRPCError:
		t.Error("got JSON-RPC error; expected a JSONRPCResponse with tool error in result")
	case *mcp.JSONRPCError:
		t.Error("got JSON-RPC error; expected a JSONRPCResponse with tool error in result")
	default:
		t.Errorf("unexpected response type: %T", resp)
	}
}
