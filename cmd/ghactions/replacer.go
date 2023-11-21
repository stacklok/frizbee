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
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/stacklok/frizbee/pkg/ghactions"
	"github.com/stacklok/frizbee/pkg/utils"
)

type replacer struct {
	ghcli         *github.Client
	dir           string
	dryRun        bool
	quiet         bool
	errOnModified bool
}

func (r *replacer) do(ctx context.Context, cmd *cobra.Command) error {
	basedir := filepath.Dir(r.dir)
	base := filepath.Base(r.dir)
	bfs := osfs.New(basedir, osfs.WithBoundOS())

	outfiles := map[string]string{}
	modified := false

	err := ghactions.TraverseGitHubActionWorkflows(bfs, base, func(path string, wflow *yaml.Node) error {
		r.logf(cmd, "Processing %s\n", path)
		m, err := ghactions.ModifyReferencesInYAML(ctx, r.ghcli, wflow)
		if err != nil {
			return fmt.Errorf("failed to process YAML file %s: %w", path, err)
		}

		modified = modified || m

		buf, err := utils.YAMLToBuffer(wflow)
		if err != nil {
			return fmt.Errorf("failed to convert YAML to buffer: %w", err)
		}

		if m {
			r.logf(cmd, "Modified %s\n", path)
			outfiles[path] = buf.String()
		}

		return nil
	})
	if err != nil {
		return err
	}

	if err := r.processOutput(cmd, bfs, outfiles); err != nil {
		return err
	}

	if r.errOnModified && modified {
		return fmt.Errorf("modified files")
	}

	return nil
}

func (r *replacer) logf(cmd *cobra.Command, format string, args ...interface{}) {
	if !r.quiet {
		fmt.Fprintf(cmd.ErrOrStderr(), format, args...)
	}
}

func (r *replacer) processOutput(cmd *cobra.Command, bfs billy.Filesystem, outfiles map[string]string) error {

	var out io.Writer

	for path, content := range outfiles {
		if r.quiet {
			out = io.Discard
		} else if r.dryRun {
			out = cmd.OutOrStdout()
		} else {
			f, err := bfs.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}

			defer func() {
				if err := f.Close(); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "failed to close file %s: %v", path, err)
				}
			}()

			out = f
		}

		_, err := fmt.Fprintf(out, "%s", content)
		if err != nil {
			return fmt.Errorf("failed to write to file %s: %w", path, err)
		}
	}

	return nil
}
