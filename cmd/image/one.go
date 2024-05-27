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

package image

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stacklok/frizbee/internal/cli"
	"github.com/stacklok/frizbee/pkg/config"
	"github.com/stacklok/frizbee/pkg/replacer"
)

// CmdOne represents the one sub-command
func CmdOne() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "one",
		Short: "Replace the tag with a digest reference",
		Long: `This utility replaces a tag of a container reference
with the corresponding digest.
	
Example:

	$ frizbee image one ghcr.io/stacklok/minder/server:latest

This will replace a tag of the container reference with the corresponding digest.

`,
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
		WithUserRegex(cliFlags.Regex)

	// Replace the passed reference
	res, err := r.ParseSingleContainerImage(cmd.Context(), ref)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), res)
	return nil
}
