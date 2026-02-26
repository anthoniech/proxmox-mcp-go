// Copyright (c) 2025 anthoniech
// Licensed under the MIT License. See LICENSE file for details.

package config

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var Cfg = Configuration{
	BindPort: 3001,
	BindHost: "0.0.0.0",
}

type LogSettings struct {
	LogFile string `yaml:"log_file"`
	Verbose bool   `yaml:"verbose"`

	AuditEnabled    bool   `yaml:"audit_log_enabled"`
	AuditFilePath   string `yaml:"audit_log_file"`
	AuditMaxSizeMB  int    `yaml:"audit_log_max_size_mb"`
	AuditMaxAgeDays int    `yaml:"audit_log_max_age_days"`
	AuditMaxBackups int    `yaml:"audit_log_max_backups"`
}

type Configuration struct {
	LogSettings `yaml:",inline"`

	BindHost string `yaml:"bind_host"`
	BindPort int    `yaml:"bind_port"`

	PVEURL     string `yaml:"pve_url"`
	PVETokenID string `yaml:"pve_token_id"`
	PVEToken   string `yaml:"pve_token"`

	MCPStdio  bool   `yaml:"mcp_stdio"`
	MCPAPIKey string `yaml:"mcp_api_key"`
}

func ResolveConfigPath(configFilename, workDir string) string {
	configFile, err := filepath.EvalSymlinks(configFilename)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Errorf("unexpected error while config file path evaluation: %s", err)
		}
		configFile = configFilename
	}
	if !filepath.IsAbs(configFile) {
		configFile = filepath.Join(workDir, configFile)
	}
	return configFile
}

func Parse(configPath string) error {
	log.Debugf("Reading config file: %s", configPath)
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Errorf("Couldn't read config file %s: %s", configPath, err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, &Cfg)
	if err != nil {
		log.Errorf("Couldn't parse config file: %s", err)
		return err
	}

	return nil
}
