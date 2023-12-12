// Copyright 2023 Stacklok, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package dockercompose provides command-line utilities to work with container images.
package dockercompose

import (
	"fmt"

	"github.com/spf13/cobra"

	intcmd "github.com/stacklok/frizbee/internal/cmd"
	"github.com/stacklok/frizbee/pkg/config"
)

// CmdCompose represents the compose yaml sub-command
func CmdCompose() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "docker-compose",
		Aliases: []string{"dockercompose", "compose"},
		Short:   "Replace container image references with checksums in docker-compose YAML files",
		Long: `This utility replaces a tag or branch reference in a container image references
with the digest hash of the referenced tag in docker-compose YAML files.

Example:

	$ frizbee docker-compose --dir . --dry-run --quiet --error
`,
		RunE:         replaceYAML,
		SilenceUsage: true,
	}

	// flags
	intcmd.DeclareYAMLReplacerFlags(cmd)

	return cmd
}

func replaceYAML(cmd *cobra.Command, _ []string) error {
	cfg, err := config.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get config from context: %w", err)
	}

	replacer, err := intcmd.NewYAMLReplacer(cmd)
	if err != nil {
		return err
	}

	return replacer.Do(cmd.Context(), cfg)
}
