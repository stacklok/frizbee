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

// Package image provides command-line utilities to work with container images.
package image

import (
	"github.com/spf13/cobra"
	"github.com/stacklok/frizbee/internal/cli"
	"github.com/stacklok/frizbee/pkg/config"
	"github.com/stacklok/frizbee/pkg/replacer"
)

// CmdContainerImage represents the containers command
func CmdContainerImage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Replace container image references with checksums",
		Long: `This utility replaces tag or branch references in yaml/yml files
with the latest commit hash of the referenced tag or branch.
	
Example:

	$ frizbee image -d <path>

This will replace all tag or branch references in all yaml files for the given directory.
`,
		RunE:         replaceCmd,
		SilenceUsage: true,
		Aliases:      []string{"containerimage", "dockercompose", "compose"}, // backwards compatibility
	}

	// flags
	cli.DeclareFrizbeeFlags(cmd, ".")

	// sub-commands
	cmd.AddCommand(CmdOne())
	cmd.AddCommand(CmdList())

	return cmd
}

func replaceCmd(cmd *cobra.Command, _ []string) error {
	// Extract the CLI flags from the cobra command
	cliFlags, err := cli.NewHelper(cmd)
	if err != nil {
		return err
	}

	// Set up the config
	cfg, err := config.FromCommand(cmd)
	if err != nil {
		return err
	}

	// Create a new replacer
	r := replacer.New(cfg).
		WithUserRegex(cliFlags.Regex)

	// Replace the tags in the directory
	res, err := r.ParseContainerImages(cmd.Context(), cliFlags.Dir)
	if err != nil {
		return err
	}

	// Process the output files
	return cliFlags.ProcessOutput(res.Processed, res.Modified)
}
