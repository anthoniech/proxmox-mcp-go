// Copyright (c) 2025 anthoniech
// Licensed under the MIT License. See LICENSE file for details.

package main

import (
	"runtime/debug"

	"github.com/anthoniech/proxmox-mcp-go/app"
)

// version will be set through ldflags.
var version = "undefined"

func main() {
	debug.SetGCPercent(10)
	app.Run(version)
}
