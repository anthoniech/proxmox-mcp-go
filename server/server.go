// Copyright (c) 2025 anthoniech
// Licensed under the MIT License. See LICENSE file for details.

package server

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	conf       *Config
	httpServer *http.Server
	echo       *echo.Echo
}

type Config struct {
	BindHost string
	BindPort int
}

func New(conf *Config) *Server {
	log.Info("Initializing server")

	s := Server{conf: conf}

	e := echo.New()

	e.File("/", "build/static/index.html")
	registerRoutes(e)

	s.echo = e

	address := net.JoinHostPort(s.conf.BindHost, strconv.Itoa(s.conf.BindPort))
	s.httpServer = &http.Server{
		Addr:        address,
		ReadTimeout: time.Second * 15,
		IdleTimeout: time.Second * 60,
		Handler:     e,
	}

	return &s
}

func (s *Server) Start() {
	address := net.JoinHostPort(s.conf.BindHost, strconv.Itoa(s.conf.BindPort))
	log.Printf("Listening on http://%s", address)

	err := s.httpServer.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func (s *Server) SetMCPHandler(h http.Handler, apiKey string) {
	handler := echo.WrapHandler(h)

	if apiKey != "" {
		s.echo.Any("/mcp", handler, bearerAuth(apiKey))
	} else {
		s.echo.Any("/mcp", handler)
	}
}

func bearerAuth(apiKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			const prefix = "Bearer "
			auth := c.Request().Header.Get("Authorization")

			if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing or invalid authorization header"})
			}

			if auth[len(prefix):] != apiKey {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid api key"})
			}

			return next(c)
		}
	}
}

func (s *Server) Close() {
	log.Info("Stopping HTTP server...")
	if s.httpServer != nil {
		_ = s.httpServer.Shutdown(context.TODO())
	}
	log.Info("Stopped HTTP server")
}
