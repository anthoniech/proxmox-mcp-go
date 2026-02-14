// Copyright (c) 2025 anthoniech
// Licensed under the MIT License. See LICENSE file for details.

package app

import (
	"os"
	"path/filepath"

	"github.com/anthoniech/proxmox-mcp-go/config"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"

	"go.elastic.co/ecslogrus"
)

type ecsFormatter struct {
	ecslogrus.Formatter
}

func (f *ecsFormatter) Format(entry *log.Entry) ([]byte, error) {
	entry.Data["service.name"] = os.Getenv("APP_NAME")
	entry.Data["service.environment"] = os.Getenv("ENVIRONMENT")

	return f.Formatter.Format(entry)
}

func configureLogger(args options) {
	configPath := config.ResolveConfigPath(appCtx.configFilename, appCtx.workDir)
	ls := config.GetLogSettings(configPath)

	if ls.LogType == "ecs" {
		log.SetFormatter(&ecsFormatter{ecslogrus.Formatter{}})
	} else {
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
		log.SetOutput(colorable.NewColorableStdout())
	}

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

	if ls.LogFile == "syslog" {
		// TODO: syslog configuration pending
		return
	}

	logFilePath := filepath.Join(appCtx.workDir, ls.LogFile)
	if filepath.IsAbs(ls.LogFile) {
		logFilePath = ls.LogFile
	}

	file, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("cannot create a log file: %s", err)
	}
	log.SetOutput(file)
}
