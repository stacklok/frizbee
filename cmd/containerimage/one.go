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

package containerimage

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	gocrv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/spf13/cobra"

	"github.com/stacklok/frizbee/pkg/containers"
)

// nolint: gochecknoglobals
var platform string

// CmdOne represents the one sub-command
func CmdOne() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "one",
		Short: "Replace the tag in container image reference",
		Long: `This utility replaces a tag or branch reference in a container image reference
with the digest hash of the referenced tag.
	
Example:

	$ frizbee containerimage one ghcr.io/stacklok/minder/helm/minder:0.20231123.829_ref.26ca90b
`,
		Args:         cobra.ExactArgs(1),
		RunE:         replaceOne,
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&platform, "platform", "p", platform, "Platform to use for digest")
	return cmd
}

func replaceOne(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	ref := args[0]

	r, err := name.ParseReference(ref)
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}

	img := r.Context().String()

	if len(platform) != 0 {
		if !strings.Contains(platform, "/") {
			return fmt.Errorf("platform must be in the format os/arch")
		}
		os, arch := strings.Split(platform, "/")[0], strings.Split(platform, "/")[1]
		containers.SetPlatform(&gocrv1.Platform{OS: os, Architecture: arch})
	}
	sum, err := containers.GetDigest(ctx, ref)
	if err != nil {
		return fmt.Errorf("failed to get checksum for action '%s': %w", ref, err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s@%s\n", img, sum)
	return nil
}
