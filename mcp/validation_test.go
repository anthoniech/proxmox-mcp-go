// Copyright (c) 2025 anthoniech
// Licensed under the MIT License. See LICENSE file for details.

package mcp //nolint:testpackage // exercises unexported validators directly

import (
	"testing"
)

func TestValidatePathSegment(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "valid node name", input: "pve1", want: "pve1"},
		{name: "valid with hyphen", input: "node-01", want: "node-01"},
		{name: "valid with underscore", input: "my_node", want: "my_node"},
		{name: "strips quotes", input: `"pve1"`, want: "pve1"},
		{name: "strips single quotes", input: "'pve1'", want: "pve1"},
		{name: "strips whitespace", input: "  pve1  ", want: "pve1"},
		{name: "empty string", input: "", wantErr: true},
		{name: "whitespace only", input: "   ", wantErr: true},
		{name: "path traversal with ../", input: "pve/../etc", wantErr: true},
		{name: "path traversal leading ../", input: "../admin", wantErr: true},
		{name: "path traversal trailing ../", input: "pve/..", wantErr: true},
		{name: "bare double dot", input: "..", wantErr: true},
		{name: "forward slash", input: "pve/other", wantErr: true},
		{name: "backslash", input: `pve\other`, wantErr: true},
		{name: "complex traversal", input: "pve/../../../etc", wantErr: true},
		{name: "vmid-style traversal", input: "100/../../admin", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validatePathSegment("node", tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePathSegment(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("validatePathSegment(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateGuestType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "qemu", input: "qemu", want: "qemu"},
		{name: "lxc", input: "lxc", want: "lxc"},
		{name: "qemu with quotes", input: `"qemu"`, want: "qemu"},
		{name: "lxc with whitespace", input: " lxc ", want: "lxc"},
		{name: "invalid type", input: "docker", wantErr: true},
		{name: "empty", input: "", wantErr: true},
		{name: "path traversal", input: "qemu/../lxc", wantErr: true},
		{name: "uppercase rejected", input: "QEMU", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateGuestType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGuestType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("validateGuestType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateStopAction(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "stop", input: "stop", want: "stop"},
		{name: "shutdown", input: "shutdown", want: "shutdown"},
		{name: "reboot", input: "reboot", want: "reboot"},
		{name: "stop with quotes", input: `"stop"`, want: "stop"},
		{name: "invalid action", input: "destroy", wantErr: true},
		{name: "empty", input: "", wantErr: true},
		{name: "path traversal", input: "stop/../delete", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateStopAction(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStopAction(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("validateStopAction(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseVMID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "valid integer", input: "100", want: "100"},
		{name: "zero", input: "0", want: "0"},
		{name: "large id", input: "99999", want: "99999"},
		{name: "with quotes", input: `"100"`, want: "100"},
		{name: "with whitespace", input: " 100 ", want: "100"},
		{name: "not a number", input: "abc", wantErr: true},
		{name: "path traversal", input: "100/../../admin", wantErr: true},
		{name: "float", input: "1.5", wantErr: true},
		{name: "empty", input: "", wantErr: true},
		{name: "negative", input: "-1", want: "-1"}, // Proxmox API will reject, but parsing allows it
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVMID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVMID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseVMID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
