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

package ghactions

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/stacklok/frizbee/internal/ghrest"
	"github.com/stacklok/frizbee/pkg/ghactions"
)

// CmdOne represents the one sub-command
func CmdOne() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "one",
		Short: "Replace the tag in GitHub Action reference",
		Long: `This utility replaces a tag or branch reference in a GitHub Action reference
with the latest commit hash of the referenced tag or branch.
	
Example:

	$ frizbee ghactions one actions/checkout@v4.1.1

This will replace the tag or branch reference for the commit hash of the
referenced tag or branch.

` + TokenHelpText + "\n",
		Args:         cobra.ExactArgs(1),
		RunE:         replaceOne,
		SilenceUsage: true,
	}

	return cmd
}

func replaceOne(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	ref := args[0]

	ghcli := ghrest.NewGhRest(os.Getenv(GitHubTokenEnvKey))

	act, ref, err := ghactions.ParseActionReference(ref)
	if err != nil {
		return fmt.Errorf("failed to parse action reference '%s': %w", ref, err)
	}

	sum, err := ghactions.GetChecksum(ctx, ghcli, act, ref)
	if err != nil {
		return fmt.Errorf("failed to get checksum for action '%s': %w", ref, err)
	}

	if ref != sum {
		fmt.Fprintf(cmd.OutOrStdout(), "%s@%s\n", act, sum)
	}

	return nil
}
