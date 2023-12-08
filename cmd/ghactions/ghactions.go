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

// Package ghactions provides command-line utilities to work with GitHub Actions.
package ghactions

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/stacklok/frizbee/internal/ghrest"
	"github.com/stacklok/frizbee/pkg/config"
	cliutils "github.com/stacklok/frizbee/pkg/utils/cli"
)

// CmdGHActions represents the ghactions command
func CmdGHActions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ghactions",
		Short: "Replace tags in GitHub Actions workflows",
		Long: `This utility replaces tag or branch references in GitHub Actions workflows
with the latest commit hash of the referenced tag or branch.
	
Example:

	$ frizbee ghactions -d .github/workflows

This will replace all tag or branch references in all GitHub Actions workflows
for the given directory.
`,
		RunE:         replace,
		SilenceUsage: true,
	}

	// flags
	cmd.Flags().StringP("dir", "d", ".github/workflows", "workflows directory")

	cliutils.DeclareReplacerFlags(cmd)

	// sub-commands
	cmd.AddCommand(CmdOne())
	cmd.AddCommand(CmdList())

	return cmd
}

func replace(cmd *cobra.Command, _ []string) error {
	dir := cmd.Flag("dir").Value.String()
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return fmt.Errorf("failed to get dry-run flag: %w", err)
	}
	errOnModified, err := cmd.Flags().GetBool("error")
	if err != nil {
		return fmt.Errorf("failed to get error flag: %w", err)
	}
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return fmt.Errorf("failed to get quiet flag: %w", err)
	}
	cfg, err := config.FromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get config from context: %w", err)
	}

	dir = cliutils.ProcessDirNameForBillyFS(dir)

	ctx := cmd.Context()

	ghcli := ghrest.NewGhRest(os.Getenv("GITHUB_TOKEN"))

	replacer := &replacer{
		Replacer: cliutils.Replacer{
			Dir:           dir,
			DryRun:        dryRun,
			Quiet:         quiet,
			ErrOnModified: errOnModified,
			Cmd:           cmd,
		},
		restIf: ghcli,
	}

	return replacer.do(ctx, cfg)
}
