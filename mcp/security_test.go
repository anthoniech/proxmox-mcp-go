// Copyright (c) 2025 anthoniech
// Licensed under the MIT License. See LICENSE file for details.

package mcp_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/server"

	mcplib "github.com/anthoniech/proxmox-mcp-go/mcp"
)

// callTool invokes a tool by name via the MCP JSON-RPC interface and returns
// the text content of the first result item.
func callTool(
	t *testing.T,
	s *mcplib.Server,
	toolName string,
	args map[string]any,
) string {
	t.Helper()

	ctx := context.Background()
	session := server.NewInProcessSession(
		server.GenerateInProcessSessionID(), nil,
	)
	ctx = s.MCPServer().WithContext(ctx, session)

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      toolName,
			"arguments": args,
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

	var rpcResp struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &rpcResp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if rpcResp.Error != nil {
		return "RPC_ERROR: " + rpcResp.Error.Message
	}

	if len(rpcResp.Result.Content) == 0 {
		t.Fatal("empty content in tool result")
	}

	return rpcResp.Result.Content[0].Text
}

// --- Destructive operation confirmation tests ---

func TestDeleteGuestRequiresConfirm(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "delete_guest", map[string]any{
		"node": "pve1",
		"vmid": "100",
	})
	if !strings.Contains(result, "destructive operation") {
		t.Errorf("expected destructive operation warning, got: %s", result)
	}
	if !strings.Contains(result, "confirm") {
		t.Errorf("expected confirm instruction, got: %s", result)
	}
}

func TestDeleteGuestConfirmFalse(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "delete_guest", map[string]any{
		"node":    "pve1",
		"vmid":    "100",
		"confirm": "false",
	})
	if !strings.Contains(result, "destructive operation") {
		t.Errorf("expected destructive operation warning with confirm=false, got: %s", result)
	}
}

func TestDeleteGuestConfirmTrue(t *testing.T) {
	s := newTestServer(t)

	// With confirm=true, should pass validation and attempt the API call
	// (which will fail because there's no real Proxmox backend).
	result := callTool(t, s, "delete_guest", map[string]any{
		"node":    "pve1",
		"vmid":    "100",
		"confirm": "true",
	})
	// Should NOT contain the destructive operation guard message
	if strings.Contains(result, "destructive operation") {
		t.Errorf("confirm=true should bypass guard, got: %s", result)
	}
}

func TestRollbackSnapshotRequiresConfirm(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "rollback_snapshot", map[string]any{
		"node":     "pve1",
		"vmid":     "100",
		"snapname": "snap1",
	})
	if !strings.Contains(result, "destructive operation") {
		t.Errorf("expected destructive operation warning, got: %s", result)
	}
}

func TestRollbackSnapshotConfirmTrue(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "rollback_snapshot", map[string]any{
		"node":     "pve1",
		"vmid":     "100",
		"snapname": "snap1",
		"confirm":  "true",
	})
	if strings.Contains(result, "destructive operation") {
		t.Errorf("confirm=true should bypass guard, got: %s", result)
	}
}

func TestRestoreBackupRequiresConfirm(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "restore_backup", map[string]any{
		"node":    "pve1",
		"vmid":    "100",
		"archive": "local:backup/vzdump-qemu-100.vma.zst",
	})
	if !strings.Contains(result, "destructive operation") {
		t.Errorf("expected destructive operation warning, got: %s", result)
	}
}

func TestRestoreBackupConfirmTrue(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "restore_backup", map[string]any{
		"node":    "pve1",
		"vmid":    "100",
		"archive": "local:backup/vzdump-qemu-100.vma.zst",
		"confirm": "true",
	})
	if strings.Contains(result, "destructive operation") {
		t.Errorf("confirm=true should bypass guard, got: %s", result)
	}
}

// --- Path injection tests ---

func TestNodePathTraversal(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "list_vms", map[string]any{
		"node": "pve/../../../etc",
	})
	if !strings.Contains(result, "invalid") {
		t.Errorf("expected validation error for path traversal, got: %s", result)
	}
}

func TestNodeWithSlash(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "get_node_status", map[string]any{
		"node": "pve/admin",
	})
	if !strings.Contains(result, "invalid") {
		t.Errorf("expected validation error for slash in node, got: %s", result)
	}
}

func TestInvalidGuestType(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "get_guest_config", map[string]any{
		"node": "pve1",
		"vmid": "100",
		"type": "docker",
	})
	if !strings.Contains(result, "invalid guest type") {
		t.Errorf("expected invalid guest type error, got: %s", result)
	}
}

func TestGuestTypeTraversal(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "start_guest", map[string]any{
		"node": "pve1",
		"vmid": "100",
		"type": "qemu/../lxc",
	})
	if !strings.Contains(result, "invalid guest type") {
		t.Errorf("expected invalid guest type error for traversal, got: %s", result)
	}
}

func TestInvalidStopAction(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "stop_guest", map[string]any{
		"node":   "pve1",
		"vmid":   "100",
		"action": "destroy",
	})
	if !strings.Contains(result, "invalid action") {
		t.Errorf("expected invalid action error, got: %s", result)
	}
}

func TestVMIDPathTraversal(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "get_guest_config", map[string]any{
		"node": "pve1",
		"vmid": "100/../../admin",
	})
	if !strings.Contains(result, "invalid vmid") {
		t.Errorf("expected invalid vmid error, got: %s", result)
	}
}

func TestVMIDNonNumeric(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "start_guest", map[string]any{
		"node": "pve1",
		"vmid": "abc",
	})
	if !strings.Contains(result, "invalid vmid") {
		t.Errorf("expected invalid vmid error, got: %s", result)
	}
}

// --- Valid guest types pass through ---

func TestValidGuestTypeQemu(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "get_guest_config", map[string]any{
		"node": "pve1",
		"vmid": "100",
		"type": "qemu",
	})
	// Should not fail on validation; will fail on API call (no backend)
	if strings.Contains(result, "invalid guest type") {
		t.Errorf("qemu should be a valid guest type, got: %s", result)
	}
}

func TestValidGuestTypeLXC(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "get_guest_config", map[string]any{
		"node": "pve1",
		"vmid": "100",
		"type": "lxc",
	})
	if strings.Contains(result, "invalid guest type") {
		t.Errorf("lxc should be a valid guest type, got: %s", result)
	}
}

// --- TLS verification config test ---

func TestNewProxmoxClientVerifySSL(t *testing.T) {
	// verifySSL=true should set InsecureSkipVerify=false
	s, err := mcplib.New(fakeURL, fakeToken, true, nil)
	if err != nil {
		t.Fatalf("New() with verifySSL=true returned error: %v", err)
	}
	if s == nil {
		t.Fatal("New() with verifySSL=true returned nil")
	}

	// verifySSL=false should set InsecureSkipVerify=true (default behavior)
	s2, err := mcplib.New(fakeURL, fakeToken, false, nil)
	if err != nil {
		t.Fatalf("New() with verifySSL=false returned error: %v", err)
	}
	if s2 == nil {
		t.Fatal("New() with verifySSL=false returned nil")
	}
}

// --- Snapshot path validation tests ---

func TestSnapnamePathTraversal(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "delete_snapshot", map[string]any{
		"node":     "pve1",
		"vmid":     "100",
		"snapname": "snap/../../../etc",
	})
	if !strings.Contains(result, "invalid") {
		t.Errorf("expected validation error for snapname traversal, got: %s", result)
	}
}

// --- Storage path validation tests ---

func TestStoragePathTraversal(t *testing.T) {
	s := newTestServer(t)

	result := callTool(t, s, "list_backups", map[string]any{
		"node":    "pve1",
		"storage": "local/../secrets",
	})
	if !strings.Contains(result, "invalid") {
		t.Errorf("expected validation error for storage traversal, got: %s", result)
	}
}

// Verify the confirm param is in the tool schema as required.
func TestDestructiveToolsHaveConfirmParam(t *testing.T) {
	s := newTestServer(t)
	tools := s.MCPServer().ListTools()

	destructiveTools := []string{"delete_guest", "rollback_snapshot", "restore_backup"}
	for _, toolName := range destructiveTools {
		tool, ok := tools[toolName]
		if !ok {
			t.Errorf("tool %q not found", toolName)
			continue
		}

		schema, ok := tool.Tool.InputSchema.Properties["confirm"]
		if !ok {
			t.Errorf("tool %q missing 'confirm' property in schema", toolName)
			continue
		}

		// Check it's a string type
		schemaMap, ok := schema.(map[string]any)
		if !ok {
			t.Errorf("tool %q confirm schema is not a map", toolName)
			continue
		}
		if schemaMap["type"] != "string" {
			t.Errorf("tool %q confirm should be type 'string', got %v", toolName, schemaMap["type"])
		}

		// Check it's required
		found := false
		for _, req := range tool.Tool.InputSchema.Required {
			if req == "confirm" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("tool %q should have 'confirm' as required parameter", toolName)
		}
	}
}

// Verify non-destructive tools do NOT have confirm param.
func TestNonDestructiveToolsNoConfirmParam(t *testing.T) {
	s := newTestServer(t)
	tools := s.MCPServer().ListTools()

	nonDestructive := []string{
		"list_vms", "list_containers", "get_guest_config",
		"start_guest", "stop_guest", "list_snapshots",
		"create_snapshot", "list_backups", "backup_guest",
	}
	for _, toolName := range nonDestructive {
		tool, ok := tools[toolName]
		if !ok {
			t.Errorf("tool %q not found", toolName)
			continue
		}

		if _, hasConfirm := tool.Tool.InputSchema.Properties["confirm"]; hasConfirm {
			t.Errorf("non-destructive tool %q should NOT have 'confirm' parameter", toolName)
		}
	}
}

// Verify that delete_guest returns isError=true when confirm is missing.
func TestDeleteGuestReturnsIsError(t *testing.T) {
	s := newTestServer(t)

	ctx := context.Background()
	session := server.NewInProcessSession(
		server.GenerateInProcessSessionID(), nil,
	)
	ctx = s.MCPServer().WithContext(ctx, session)

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      10,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "delete_guest",
			"arguments": map[string]any{
				"node": "pve1",
				"vmid": "100",
			},
		},
	}
	msg, _ := json.Marshal(req)
	resp := s.MCPServer().HandleMessage(ctx, msg)

	data, _ := json.Marshal(resp)
	var rpcResp struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &rpcResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !rpcResp.Result.IsError {
		t.Error("expected isError=true when confirm is missing")
	}
	if len(rpcResp.Result.Content) > 0 &&
		!strings.Contains(rpcResp.Result.Content[0].Text, "destructive") {
		t.Errorf("expected destructive warning, got: %s", rpcResp.Result.Content[0].Text)
	}
}
