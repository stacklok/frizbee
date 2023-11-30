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

// Package cli provides utilities for frizbee's CLI.
package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/spf13/cobra"
)

// DeclareReplacerFlags declares the flags common to all replacer commands.
// Note that `dir` is not declared here because it is command-specific.
func DeclareReplacerFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("dry-run", "n", false, "don't modify files")
	cmd.Flags().BoolP("quiet", "q", false, "don't print anything")
	cmd.Flags().BoolP("error", "e", false, "exit with error code if any file is modified")
}

// Replacer is a common struct for implementing a CLI command that replaces
// files.
type Replacer struct {
	Dir           string
	DryRun        bool
	Quiet         bool
	ErrOnModified bool
	Cmd           *cobra.Command
}

// Logf logs the given message to the given command's stderr if the command is
// not quiet.
func (r *Replacer) Logf(format string, args ...interface{}) {
	if !r.Quiet {
		fmt.Fprintf(r.Cmd.ErrOrStderr(), format, args...)
	}
}

// ProcessOutput processes the given output files.
// If the command is quiet, the output is discarded.
// If the command is a dry run, the output is written to the command's stdout.
// Otherwise, the output is written to the given filesystem.
func (r *Replacer) ProcessOutput(bfs billy.Filesystem, outfiles map[string]string) error {

	var out io.Writer

	for path, content := range outfiles {
		if r.Quiet {
			out = io.Discard
		} else if r.DryRun {
			out = r.Cmd.OutOrStdout()
		} else {
			f, err := bfs.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}

			defer func() {
				if err := f.Close(); err != nil {
					fmt.Fprintf(r.Cmd.ErrOrStderr(), "failed to close file %s: %v", path, err)
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
