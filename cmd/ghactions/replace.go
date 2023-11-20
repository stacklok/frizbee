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
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	billyutil "github.com/go-git/go-billy/v5/util"
	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/stacklok/frisbee/pkg/ghactions"
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

	err := billyutil.Walk(bfs, base, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("failed to walk path %s: %v\n", path, err)
			return nil
		}

		if shouldSkipFile(info) {
			fmt.Printf("skipping file %s\n", path)
			return nil
		}

		f, err := bfs.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer f.Close()

		dec := yaml.NewDecoder(f)

		var wflow yaml.Node
		if err := dec.Decode(&wflow); err != nil {
			return fmt.Errorf("failed to decode file %s: %w", path, err)
		}

		if err := traverseYAML(ctx, ghcli, &wflow); err != nil {
			return fmt.Errorf("failed to process YAML file %s: %w", path, err)
		}

		buf := strings.Builder{}
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		enc.Encode(&wflow)
		enc.Close()

		outfiles[path] = buf.String()

		return nil
	})
	if err != nil {
		return err
	}

	processOutput(cmd, bfs, outfiles)

	return nil
}

func traverseYAML(ctx context.Context, ghcli *github.Client, node *yaml.Node) error {
	// `uses` will be immediately before the action
	// name in the YAML `Content` array. We use a toggle
	// to track if we've found `uses` and then look for
	// the next node.
	foundUses := false

	for _, v := range node.Content {
		if v.Value == "uses" {
			foundUses = true
			continue
		}

		if foundUses {
			foundUses = false

			// If the value is a local path, skip it
			if ghactions.IsLocal(v.Value) {
				continue
			}

			act, ref, err := ghactions.ParseActionReference(v.Value)
			if err != nil {
				return fmt.Errorf("failed to parse action reference '%s': %w", v.Value, err)
			}

			sum, err := ghactions.GetChecksum(ctx, ghcli, act, ref)
			if err != nil {
				return fmt.Errorf("failed to get checksum for action '%s': %w", v.Value, err)
			}

			if ref != sum {
				v.SetString(fmt.Sprintf("%s@%s", act, sum))
				v.LineComment = ref
			}
			continue
		}

		// Otherwise recursively look more
		if err := traverseYAML(ctx, ghcli, v); err != nil {
			return err
		}
	}
	return nil
}

func shouldSkipFile(info fs.FileInfo) bool {
	// skip if not a file
	if info.IsDir() {
		return true
	}

	// skip if not a .yml or .yaml file
	if !strings.HasSuffix(info.Name(), ".yml") && !strings.HasSuffix(info.Name(), ".yaml") {
		return true
	}

	return false
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
