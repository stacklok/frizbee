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

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync/atomic"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/stacklok/frizbee/pkg/config"
	"github.com/stacklok/frizbee/pkg/containers"
	ferrors "github.com/stacklok/frizbee/pkg/errors"
	"github.com/stacklok/frizbee/pkg/utils"
	cliutils "github.com/stacklok/frizbee/pkg/utils/cli"
)

// DeclareYAMLReplacerFlags declares the flags for the YAML replacer
func DeclareYAMLReplacerFlags(cli *cobra.Command) {
	cli.Flags().StringP("dir", "d", ".", "manifests file or directory")

	cliutils.DeclareReplacerFlags(cli)
}

// YAMLReplacer replaces container image references in YAML files
type YAMLReplacer struct {
	cliutils.Replacer
	ImageRegex string
}

// WithImageRegex sets the image regex
func WithImageRegex(regex string) func(*YAMLReplacer) {
	return func(r *YAMLReplacer) {
		r.ImageRegex = regex
	}
}

// NewYAMLReplacer creates a new YAMLReplacer from the given
// command-line arguments and options
func NewYAMLReplacer(cli *cobra.Command, opts ...func(*YAMLReplacer)) (*YAMLReplacer, error) {
	dir := cli.Flag("dir").Value.String()
	dryRun, err := cli.Flags().GetBool("dry-run")
	if err != nil {
		return nil, fmt.Errorf("failed to get dry-run flag: %w", err)
	}
	errOnModified, err := cli.Flags().GetBool("error")
	if err != nil {
		return nil, fmt.Errorf("failed to get error flag: %w", err)
	}
	quiet, err := cli.Flags().GetBool("quiet")
	if err != nil {
		return nil, fmt.Errorf("failed to get quiet flag: %w", err)
	}

	dir = cliutils.ProcessDirNameForBillyFS(dir)

	r := &YAMLReplacer{
		Replacer: cliutils.Replacer{
			Dir:           dir,
			DryRun:        dryRun,
			Quiet:         quiet,
			ErrOnModified: errOnModified,
			Cmd:           cli,
		},
		ImageRegex: "image",
	}
	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

// Do runs the YAMLReplacer
func (r *YAMLReplacer) Do(ctx context.Context, _ *config.Config) error {
	basedir := filepath.Dir(r.Dir)
	base := filepath.Base(r.Dir)
	// NOTE: For some reason using boundfs causes a panic when trying to open a file.
	// I instead falled back to chroot which is the default.
	bfs := osfs.New(basedir)

	outfiles := map[string]string{}

	var modified atomic.Bool
	modified.Store(false)

	var eg errgroup.Group
	cache := utils.NewRefCacher()

	err := utils.Traverse(bfs, base, func(path string, info fs.FileInfo) error {
		eg.Go(func() error {
			if !utils.IsYAMLFile(info) {
				return nil
			}

			f, err := bfs.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}

			// nolint:errcheck // ignore error
			defer f.Close()

			r.Logf("Processing %s\n", path)

			buf := bytes.Buffer{}
			m, err := containers.ReplaceReferenceFromYAMLWithCache(ctx, r.ImageRegex, f, &buf, cache)
			if err != nil {
				return fmt.Errorf("failed to process YAML file %s: %w", path, err)
			}

			modified.Store(modified.Load() || m)

			if m {
				r.Logf("Modified %s\n", path)
				outfiles[path] = buf.String()
			}

			return nil
		})

		return nil
	})
	if err != nil {
		return err
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	if err := r.ProcessOutput(bfs, outfiles); err != nil {
		return err
	}

	if r.ErrOnModified && modified.Load() {
		return ferrors.ErrModifiedFiles
	}

	return nil
}
