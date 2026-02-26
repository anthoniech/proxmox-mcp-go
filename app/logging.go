// Copyright (c) 2025 anthoniech
// Licensed under the MIT License. See LICENSE file for details.

package app

import (
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/anthoniech/proxmox-mcp-go/config"

	"gopkg.in/natefinch/lumberjack.v2"
)

// AuditLogger is a dedicated logger for Proxmox API call audit entries.
var AuditLogger *log.Logger

func configureLogger(args options) {
	ls := config.Cfg.LogSettings

	if args.verbose {
		ls.Verbose = true
	}
	if args.logFile != "" {
		ls.LogFile = args.logFile
	}

	if ls.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	if ls.LogFile == "" {
		return
	}

	logFilePath := resolveFilePath(ls.LogFile)

	file, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("cannot create a log file: %s", err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, file))
}

func resolveFilePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(appCtx.workDir, path)
}

func configureAuditLogger() {
	ls := config.Cfg.LogSettings
	if !ls.AuditEnabled {
		return
	}

	filePath := resolveFilePath(ls.AuditFilePath)
	if filePath == "" {
		filePath = resolveFilePath("logs/audit.log")
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0750); err != nil {
		log.Errorf("cannot create audit log directory: %s", err)
		return
	}

	maxSize := ls.AuditMaxSizeMB
	if maxSize <= 0 {
		maxSize = 100
	}
	maxAge := ls.AuditMaxAgeDays
	if maxAge <= 0 {
		maxAge = 30
	}
	maxBackups := ls.AuditMaxBackups
	if maxBackups <= 0 {
		maxBackups = 5
	}

	rotator := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		MaxBackups: maxBackups,
		Compress:   true,
	}

	AuditLogger = log.New()
	AuditLogger.SetFormatter(&log.JSONFormatter{})
	AuditLogger.SetOutput(io.MultiWriter(os.Stdout, rotator))
	AuditLogger.SetLevel(log.InfoLevel)

	log.Infof("Audit logger configured, writing to %s", filePath)
}
