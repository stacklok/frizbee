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
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/stacklok/boomerang/pkg/ghactions"
	"github.com/stacklok/boomerang/pkg/utils"
)

// replaceCmd represents the replace command
var replaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "Replace tags in GitHub Actions workflows",
	RunE:  replace,
}

func init() {
	GHActionsCmd.AddCommand(replaceCmd)

	replaceCmd.Flags().StringP("dir", "d", ".github/workflows", "workflows directory")
	replaceCmd.Flags().BoolP("dry-run", "n", false, "dry run")
}

func replace(cmd *cobra.Command, args []string) error {
	dir := cmd.Flag("dir").Value.String()

	ctx := cmd.Context()

	ghcli := github.NewClient(nil)

	tok := os.Getenv("GITHUB_TOKEN")
	if tok != "" {
		ghcli = ghcli.WithAuthToken(tok)
	}

	basedir := filepath.Dir(dir)
	base := filepath.Base(dir)
	bfs := osfs.New(basedir, osfs.WithBoundOS())

	outfiles := map[string]string{}

	err := ghactions.TraverseGitHubActionWorkflows(bfs, base, func(path string, wflow *yaml.Node) error {
		fmt.Fprintf(cmd.ErrOrStderr(), "Processing %s\n", path)
		if _, err := ghactions.ModifyReferencesInYAML(ctx, ghcli, wflow); err != nil {
			return fmt.Errorf("failed to process YAML file %s: %w", path, err)
		}

		buf, err := utils.YAMLToBuffer(wflow)
		if err != nil {
			return fmt.Errorf("failed to convert YAML to buffer: %w", err)
		}

		outfiles[path] = buf.String()

		return nil
	})
	if err != nil {
		return err
	}

	processOutput(cmd, bfs, outfiles)

	return nil
}

func processOutput(cmd *cobra.Command, bfs billy.Filesystem, outfiles map[string]string) error {
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return fmt.Errorf("failed to get dry-run flag: %w", err)
	}

	var out io.Writer

	for path, content := range outfiles {
		if dryRun {
			out = cmd.OutOrStdout()
		} else {
			f, err := bfs.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}

			defer f.Close()

			out = f
		}

		_, err := fmt.Fprintf(out, "%s", content)
		if err != nil {
			return fmt.Errorf("failed to write to file %s: %w", path, err)
		}
	}

	return nil
}
