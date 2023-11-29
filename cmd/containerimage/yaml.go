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

// Package containerimage provides command-line utilities to work with container images.
package containerimage

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/stacklok/frizbee/pkg/config"
	cliutils "github.com/stacklok/frizbee/pkg/utils/cli"
)

// CmdYAML represents the yaml sub-command
func CmdYAML() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yaml",
		Short: "Replace container image references with checksums in YAML files",
		Long: `This utility replaces a tag or branch reference in a container image references
with the digest hash of the referenced tag in YAML files.

Example:

	$ frizbee containerimage yaml --dir . --dry-run --quiet --error
`,
		RunE:         replaceYAML,
		SilenceUsage: true,
	}

	// flags
	cmd.Flags().StringP("dir", "d", ".", "workflows directory")
	cmd.Flags().StringP("image-regex", "i", "image", "regex to match container image references")

	cliutils.DeclareReplacerFlags(cmd)

	return cmd
}

func replaceYAML(cmd *cobra.Command, _ []string) error {
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
	ir, err := cmd.Flags().GetString("image-regex")
	if err != nil {
		return fmt.Errorf("failed to get image-regex flag: %w", err)
	}

	dir = cliutils.ProcessDirNameForBillyFS(dir)

	ctx := cmd.Context()

	replacer := &yamlReplacer{
		Replacer: cliutils.Replacer{
			Dir:           dir,
			DryRun:        dryRun,
			Quiet:         quiet,
			ErrOnModified: errOnModified,
			Cmd:           cmd,
		},
		imageRegex: ir,
	}

	return replacer.do(ctx, cfg)
}
