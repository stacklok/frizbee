//
// Copyright 2023 Stacklok, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package config provides the frizbee configuration.
package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"gopkg.in/yaml.v3"
)

type contextConfigKey struct{}

// ContextConfigKey is the context key for the configuration.
// nolint:gochecknoglobals // this is a context key
var ContextConfigKey = contextConfigKey{}

// FromContext returns the configuration from the context.
func FromContext(ctx context.Context) (*Config, error) {
	cfg, ok := ctx.Value(ContextConfigKey).(*Config)
	if !ok {
		return nil, fmt.Errorf("failed to get config from context")
	}

	return cfg, nil
}

// Config is the frizbee configuration.
type Config struct {
	GHActions GHActions `yaml:"ghactions" mapstructure:"ghactions"`
}

// GHActions is the GitHub Actions configuration.
type GHActions struct {
	Filter `yaml:",inline" mapstructure:",inline"`
}

// Filter is a common configuration for filtering out patterns.
type Filter struct {
	// Exclude is a list of patterns to exclude.
	Exclude []string `yaml:"exclude" mapstructure:"exclude"`
}

// ParseConfigFile parses a configuration file.
func ParseConfigFile(configfile string) (*Config, error) {
	bfs := osfs.New(".")
	return ParseConfigFileFromFS(bfs, configfile)
}

// ParseConfigFileFromFS parses a configuration file from a filesystem.
func ParseConfigFileFromFS(fs billy.Filesystem, configfile string) (*Config, error) {
	cfg := &Config{}
	cleancfgfile := filepath.Clean(configfile)
	cfgF, err := fs.Open(cleancfgfile)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}

		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	// nolint:errcheck // we don't care about the error here
	defer cfgF.Close()

	dec := yaml.NewDecoder(cfgF)

	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return cfg, nil
}
