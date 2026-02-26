// Copyright (c) 2025 anthoniech
// Licensed under the MIT License. See LICENSE file for details.

package app

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"

	"github.com/anthoniech/proxmox-mcp-go/config"
	"github.com/anthoniech/proxmox-mcp-go/mcp"
	"github.com/anthoniech/proxmox-mcp-go/server"
)

var versionString = "dev"

type options struct {
	verbose        bool
	configFilename string
	logFile        string
	workDir        string
}

var appCtx struct {
	web              *server.Server
	mcpServer        *mcp.Server
	configFilename   string
	workDir          string
	appSignalChannel chan os.Signal
}

func Run(version string) {
	versionString = version
	args := loadOptions()

	appCtx.appSignalChannel = make(chan os.Signal, 1)
	signal.Notify(appCtx.appSignalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		for {
			sig := <-appCtx.appSignalChannel
			log.Infof("Received signal '%s'", sig)
			switch sig {
			case syscall.SIGHUP:
			default:
				cleanup()
				os.Exit(0)
			}
		}
	}()

	run(args)
}

func loadOptions() options {
	o := options{}

	flag.StringVar(&o.configFilename, "config", "", "Path to the config file")
	flag.StringVar(&o.workDir, "w", "", "Path to the working directory")
	flag.StringVarP(
		&o.logFile,
		"log",
		"l",
		"",
		"Path to log file. If empty: write to stdout",
	)
	verbose := flag.BoolP("verbose", "v", false, "verbose output")
	help := flag.BoolP("help", "h", false, "Print this help")

	flag.CommandLine.SortFlags = false
	flag.Parse()

	if *help {
		fmt.Printf("Usage:\n\n")                   //nolint:forbidigo // Printf is appropriate for CLI help output
		fmt.Printf("%s [options]\n\n", os.Args[0]) //nolint:forbidigo // Printf is appropriate for CLI help output
		fmt.Printf("Options:\n")                   //nolint:forbidigo // Printf is appropriate for CLI help output
		flag.PrintDefaults()
		os.Exit(64)
	}

	if *verbose {
		o.verbose = true
	}

	return o
}

func run(args options) {
	if args.configFilename != "" {
		appCtx.configFilename = args.configFilename
	} else {
		appCtx.configFilename = "config.yaml"
	}

	initWorkingDir(args)

	configPath := config.ResolveConfigPath(appCtx.configFilename, appCtx.workDir)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Error("Configuration file not found, exiting")
		os.Exit(1)
	}

	err := config.Parse(configPath)
	if err != nil {
		log.Error("Failed to parse configuration, exiting")
		os.Exit(1)
	}

	configureLogger(args)
	configureAuditLogger()

	log.Printf("proxmox-mcp-go, version %s, arch %s %s", versionString, runtime.GOOS, runtime.GOARCH)
	log.Debugf("Current working directory is %s", appCtx.workDir)

	webConf := server.Config{
		BindHost: config.Cfg.BindHost,
		BindPort: config.Cfg.BindPort,
	}
	appCtx.web = server.New(&webConf)
	if appCtx.web == nil {
		log.Panicf("Can't initialize server")
	}

	if config.Cfg.PVEURL != "" && config.Cfg.PVETokenID != "" && config.Cfg.PVEToken != "" {
		pveToken := config.Cfg.PVETokenID + "=" + config.Cfg.PVEToken
		mcpSrv, err := mcp.New(config.Cfg.PVEURL, pveToken, AuditLogger)
		if err != nil {
			log.Warnf("Failed to initialize MCP server: %v", err)
		} else {
			appCtx.mcpServer = mcpSrv
			appCtx.web.SetMCPHandler(mcpSrv.Handler(), config.Cfg.MCPAPIKey)

			if config.Cfg.MCPStdio {
				go appCtx.mcpServer.Start()
			}
		}
	} else {
		log.Warn("PVE configuration not set, MCP server will not start")
	}

	appCtx.web.Start()

	select {}
}

func initWorkingDir(args options) {
	execPath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	if args.workDir != "" {
		appCtx.workDir = args.workDir
	} else {
		appCtx.workDir = filepath.Dir(execPath)
	}
}

func cleanup() {
	log.Info("Stopping proxmox-mcp-go")

	if appCtx.mcpServer != nil {
		appCtx.mcpServer.Close()
		appCtx.mcpServer = nil
	}

	if appCtx.web != nil {
		appCtx.web.Close()
		appCtx.web = nil
	}

	log.Info("Shutting down...")
}
