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

package replacer

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stacklok/frizbee/internal/cli"
	"github.com/stacklok/frizbee/pkg/config"
	"github.com/stacklok/frizbee/pkg/interfaces"
	"github.com/stacklok/frizbee/pkg/replacer/actions"
	"github.com/stacklok/frizbee/pkg/replacer/image"
)

func TestReplacer_ParseContainerImageString(t *testing.T) {
	t.Parallel()

	type args struct {
		refstr string
	}
	tests := []struct {
		name    string
		args    args
		want    *interfaces.EntityRef
		wantErr bool
	}{
		{
			name: "dockerfile - tag",
			args: args{
				refstr: "FROM golang:1.22.2",
			},
			want: &interfaces.EntityRef{
				Name:   "index.docker.io/library/golang",
				Ref:    "sha256:d5302d40dc5fbbf38ec472d1848a9d2391a13f93293a6a5b0b87c99dc0eaa6ae",
				Type:   image.ReferenceType,
				Tag:    "1.22.2",
				Prefix: "FROM ",
			},
			wantErr: false,
		},
		{
			name: "dockerfile - already by digest",
			args: args{
				refstr: "FROM golang:1.22.2@sha256:aca60c1f21de99aa3a34e653f0cdc8c8ea8fe6480359229809d5bcb974f599ec",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "dockerfile - scratch",
			args: args{
				refstr: "FROM scratch",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid 1",
			args: args{
				refstr: "ghcr.io/stacklok/minder/helm/minder:0.20231123.829_ref.26ca90b",
			},
			want: &interfaces.EntityRef{
				Name:   "ghcr.io/stacklok/minder/helm/minder",
				Ref:    "sha256:a29f8a8d28f0af7f70a4b3dd3e33c8c8cc5cf9e88e802e2700cf272a0b6140ec",
				Type:   image.ReferenceType,
				Tag:    "0.20231123.829_ref.26ca90b",
				Prefix: "",
			},
			wantErr: false,
		},
		{
			name: "valid 2",
			args: args{
				refstr: "devopsfaith/krakend:2.5.0",
			},
			want: &interfaces.EntityRef{
				Name:   "index.docker.io/devopsfaith/krakend",
				Ref:    "sha256:6a3c8e5e1a4948042bfb364ed6471e16b4a26d0afb6c3c01ebcb88b3fa551036",
				Type:   image.ReferenceType,
				Tag:    "2.5.0",
				Prefix: "",
			},
			wantErr: false,
		},
		{
			name: "invalid ref string",
			args: args{
				refstr: "ghcr.io/stacklok/minder/helm/minder!",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "nonexistent container in nonexistent registry",
			args: args{
				refstr: "beeeeer.io/ipa/toppling-goliath/king-sue:1.0.0",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			r := NewImageReplacer(&config.Config{})
			got, err := r.ParseString(ctx, tt.args.refstr)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReplacer_ParseGitHubActionString(t *testing.T) {
	t.Parallel()

	type args struct {
		action string
	}
	tests := []struct {
		name    string
		args    args
		want    *interfaces.EntityRef
		wantErr bool
	}{
		{
			name: "action using a container via docker://avtodev/markdown-lint:v1",
			args: args{
				action: "uses: docker://avtodev/markdown-lint:v1",
			},
			want: &interfaces.EntityRef{
				Name:   "index.docker.io/avtodev/markdown-lint",
				Ref:    "sha256:6aeedc2f49138ce7a1cd0adffc1b1c0321b841dc2102408967d9301c031949ee",
				Type:   image.ReferenceType,
				Tag:    "v1",
				Prefix: "uses: docker://",
			},
			wantErr: false,
		},
		{
			name: "actions/checkout with v4.1.1",
			args: args{
				action: "actions/checkout@v4.1.1",
			},
			want: &interfaces.EntityRef{
				Name:   "actions/checkout",
				Ref:    "b4ffde65f46336ab88eb53be808477a3936bae11",
				Type:   actions.ReferenceType,
				Tag:    "v4.1.1",
				Prefix: "",
			},
			wantErr: false,
		},
		{
			name: "actions/checkout with v3.6.0",
			args: args{
				action: "uses: actions/checkout@v3.6.0",
			},
			want: &interfaces.EntityRef{
				Name:   "actions/checkout",
				Ref:    "f43a0e5ff2bd294095638e18286ca9a3d1956744",
				Type:   actions.ReferenceType,
				Tag:    "v3.6.0",
				Prefix: "uses: ",
			},
			wantErr: false,
		},
		{
			name: "actions/checkout with checksum returns checksum",
			args: args{
				action: "actions/checkout@1d96c772d19495a3b5c517cd2bc0cb401ea0529f",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "aquasecurity/trivy-action with 0.14.0",
			args: args{
				action: "aquasecurity/trivy-action@0.14.0",
			},
			want: &interfaces.EntityRef{
				Name:   "aquasecurity/trivy-action",
				Ref:    "2b6a709cf9c4025c5438138008beaddbb02086f0",
				Type:   actions.ReferenceType,
				Tag:    "0.14.0",
				Prefix: "",
			},
			wantErr: false,
		},
		{
			name: "aquasecurity/trivy-action with branch returns checksum",
			args: args{
				action: "aquasecurity/trivy-action@bump-trivy",
			},
			want: &interfaces.EntityRef{
				Name:   "aquasecurity/trivy-action",
				Ref:    "fb5e1b36be448e92ca98648c661bd7e9da1f1317",
				Type:   actions.ReferenceType,
				Tag:    "bump-trivy",
				Prefix: "",
			},
			wantErr: false,
		},
		{
			name: "actions/checkout with invalid tag returns error",
			args: args{
				action: "actions/checkout@v4.1.1.1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "actions/checkout with invalid action returns error",
			args: args{
				action: "invalid-action@v4.1.1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "actions/checkout with empty action returns error",
			args: args{
				action: "@v4.1.1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "actions/checkout with empty tag returns error",
			args: args{
				action: "actions/checkout",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "bufbuild/buf-setup-action with v1 is an array",
			args: args{
				action: "bufbuild/buf-setup-action@v1",
			},
			want: &interfaces.EntityRef{
				Name:   "bufbuild/buf-setup-action",
				Ref:    "f0475db2e1b1b2e8d121066b59dfb7f7bd6c4dc4",
				Type:   actions.ReferenceType,
				Tag:    "v1",
				Prefix: "",
			},
		},
		{
			name: "anchore/sbom-action/download-syft with a sub-action works",
			args: args{
				action: "anchore/sbom-action/download-syft@v0.14.3",
			},
			want: &interfaces.EntityRef{
				Name:   "anchore/sbom-action/download-syft",
				Ref:    "78fc58e266e87a38d4194b2137a3d4e9bcaf7ca1",
				Type:   actions.ReferenceType,
				Tag:    "v0.14.3",
				Prefix: "",
			},
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := NewActionsReplacer(&config.Config{}).WithGitHubClient(os.Getenv("GITHUB_TOKEN"))
			got, err := r.ParseString(context.Background(), tt.args.action)
			if tt.wantErr {
				require.Error(t, err, "Wanted error, got none")
				require.Empty(t, got, "Wanted empty string, got %v", got)
				return
			}
			require.NoError(t, err, "Wanted no error, got %v", err)
			require.Equal(t, tt.want, got, "Wanted %v, got %v", tt.want, got)
		})
	}
}

func TestReplacer_ParseContainerImagesInFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		before   string
		expected string
		modified bool
		wantErr  bool
	}{
		{
			name: "Replace image reference",
			before: `
version: v1
services:
  - name: kube-apiserver
    image: registry.k8s.io/kube-apiserver:v1.20.0
  - name: kube-controller-manager
    image: registry.k8s.io/kube-controller-manager:v1.15.0
  - name: minder-app
    image: minder:latest
`,
			expected: `
version: v1
services:
  - name: kube-apiserver
    image: registry.k8s.io/kube-apiserver@sha256:8b8125d7a6e4225b08f04f65ca947b27d0cc86380bf09fab890cc80408230114 # v1.20.0
  - name: kube-controller-manager
    image: registry.k8s.io/kube-controller-manager@sha256:835f32a5cdb30e86f35675dd91f9c7df01d48359ab8b51c1df866a2c7ea2e870 # v1.15.0
  - name: minder-app
    image: minder:latest
`,
			modified: true,
		},
		// Add more test cases as needed
	}

	// Define a regular expression to match YAML tags containing "image"
	for _, tt := range testCases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			r := NewImageReplacer(&config.Config{})
			modified, newContent, err := r.ParseFile(ctx, strings.NewReader(tt.before))

			if tt.modified {
				assert.True(t, modified)
				assert.NotEmpty(t, newContent)
			} else {
				assert.False(t, modified)
				assert.Empty(t, newContent)
			}

			if tt.wantErr {
				assert.False(t, modified)
				assert.Empty(t, newContent)
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, newContent)
		})
	}
}

func TestReplacer_ParseGitHubActionsInFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		before   string
		expected string
		modified bool
		wantErr  bool
	}{
		{
			name: "Replace image reference",
			before: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./minder/server.yml # this should not be replaced
      - uses: actions/checkout@v2
      - uses: xt0rted/markdownlint-problem-matcher@v1
      - name: "Run Markdown linter"
        uses: docker://avtodev/markdown-lint:v1
        with:
          args: src/*.md
`,
			expected: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./minder/server.yml # this should not be replaced
      - uses: actions/checkout@ee0669bd1cc54295c223e0bb666b733df41de1c5 # v2
      - uses: xt0rted/markdownlint-problem-matcher@c17ca40d1376f60aba7e7d38a8674a3f22f7f5b0 # v1
      - name: "Run Markdown linter"
        uses: docker://index.docker.io/avtodev/markdown-lint@sha256:6aeedc2f49138ce7a1cd0adffc1b1c0321b841dc2102408967d9301c031949ee # v1
        with:
          args: src/*.md
`,
			modified: true,
		},
		// Add more test cases as needed
	}

	// Define a regular expression to match YAML tags containing "image"
	for _, tt := range testCases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			r := NewActionsReplacer(&config.Config{}).WithGitHubClient(os.Getenv(cli.GitHubTokenEnvKey))
			modified, newContent, err := r.ParseFile(ctx, strings.NewReader(tt.before))

			if tt.modified {
				assert.True(t, modified)
				assert.NotEmpty(t, newContent)
			} else {
				assert.False(t, modified)
				assert.Empty(t, newContent)
			}

			if tt.wantErr {
				assert.False(t, modified)
				assert.Empty(t, newContent)
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, newContent)
		})
	}
}
