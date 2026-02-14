// Copyright (c) 2025 anthoniech
// Licensed under the MIT License. See LICENSE file for details.

package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func registerRoutes(e *echo.Echo) {
	g := e.Group("/api/v1")

	g.GET("/", defaultPage)
	g.GET("/health", healthCheck)
}

func defaultPage(c echo.Context) error {
	return c.String(http.StatusOK, "Proxmox MCP API - v1")
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, "Proxmox MCP API Service is healthy!")
}
