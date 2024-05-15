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

package action

import (
	"fmt"
	"github.com/stacklok/frizbee/internal/cli"
	"github.com/stacklok/frizbee/pkg/config"
	"github.com/stacklok/frizbee/pkg/replacer"
	"os"

	"github.com/spf13/cobra"
)

// CmdOne represents the one sub-command
func CmdOne() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "one",
		Short: "Replace the tag in GitHub Action reference",
		Long: `This utility replaces a tag or branch reference in a GitHub Action reference
with the latest commit hash of the referenced tag or branch.
	
Example:

	$ frizbee action one actions/checkout@v4.1.1

This will replace the tag or branch reference for the commit hash of the
referenced tag or branch.

` + cli.TokenHelpText + "\n",
		Args:         cobra.ExactArgs(1),
		RunE:         replaceOne,
		SilenceUsage: true,
	}
	cli.DeclareFrizbeeFlags(cmd, "")

	return cmd
}

func replaceOne(cmd *cobra.Command, args []string) error {
	ref := args[0]

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

	// Replace the passed reference
	res, err := r.ParseSingleGitHubAction(cmd.Context(), ref)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), res)
	return nil
}
