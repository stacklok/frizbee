//
// Copyright 2024 Stacklok, Inc.
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

// Package action provides command-line utilities to work with GitHub Actions.
package action

import (
	"fmt"
	"github.com/stacklok/frizbee/internal/cli"
	"github.com/stacklok/frizbee/pkg/config"
	"github.com/stacklok/frizbee/pkg/replacer"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// CmdGHActions represents the actions command
func CmdGHActions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "action",
		Short: "Replace tags in GitHub Actions workflows",
		Long: `This utility replaces tag or branch references in GitHub Actions workflows
with the latest commit hash of the referenced tag or branch.
	
Example:

	$ frizbee action <.github/workflows> or <actions/checkout@v4>

This will replace all tag or branch references in all GitHub Actions workflows
for the given directory. Supports both directories and single references.

` + cli.TokenHelpText + "\n",
		Aliases:      []string{"ghactions"}, // backwards compatibility
		RunE:         replaceCmd,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
	}

	// flags
	cli.DeclareFrizbeeFlags(cmd, false)

	// sub-commands
	cmd.AddCommand(CmdList())

	return cmd
}

func replaceCmd(cmd *cobra.Command, args []string) error {
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
		WithUserRegex(cliFlags.Regex).
		WithGitHubClient(os.Getenv(cli.GitHubTokenEnvKey))

	if cli.IsPath(args[0]) {
		dir := filepath.Clean(args[0])
		// Replace the tags in the given directory
		res, err := r.ParseGitHubActions(cmd.Context(), dir)
		if err != nil {
			return err
		}
		// Process the output files
		return cliFlags.ProcessOutput(dir, res.Processed, res.Modified)
	} else {
		// Replace the passed reference
		res, err := r.ParseSingleGitHubAction(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), res)
		return nil
	}
}
