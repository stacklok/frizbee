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

	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
)

// CmdGHActions represents the ghactions command
func CmdGHActions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ghactions",
		Short: "Replace tags in GitHub Actions workflows",
		Long: `This utility replaces tag or branch references in GitHub Actions workflows
with the latest commit hash of the referenced tag or branch.
	
Example:

	$ boomerang ghactions -d .github/workflows

This will replace all tag or branch references in all GitHub Actions workflows
for the given directory.
`,
		RunE:         replace,
		SilenceUsage: true,
	}

	// flags
	cmd.Flags().StringP("dir", "d", ".github/workflows", "workflows directory")
	cmd.Flags().BoolP("dry-run", "n", false, "don't modify files")
	cmd.Flags().BoolP("quiet", "q", false, "don't print anything")
	cmd.Flags().BoolP("error", "e", false, "exit with error code if any file is modified")

	// sub-commands
	cmd.AddCommand(CmdOne())

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

	// remove trailing / from dir. This doesn't play well with
	// the go-billy filesystem and walker we use.
	if dir[len(dir)-1] == '/' {
		dir = dir[:len(dir)-1]
	}

	ctx := cmd.Context()

	ghcli := github.NewClient(nil)

	tok := os.Getenv("GITHUB_TOKEN")
	if tok != "" {
		ghcli = ghcli.WithAuthToken(tok)
	}

	replacer := &replacer{
		ghcli:         ghcli,
		dir:           dir,
		dryRun:        dryRun,
		quiet:         quiet,
		errOnModified: errOnModified,
	}

	return replacer.do(ctx, cmd)
}
