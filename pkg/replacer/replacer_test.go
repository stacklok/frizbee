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
	"github.com/stacklok/frizbee/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestReplacer_ParseSingleContainerImage(t *testing.T) {
	t.Parallel()

	type args struct {
		refstr string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid 1",
			args: args{
				refstr: "ghcr.io/stacklok/minder/helm/minder:0.20231123.829_ref.26ca90b",
			},
			want: "ghcr.io/stacklok/minder/helm/minder@sha256:a29f8a8d28f0af7f70a4b3dd3e33c8c8cc5cf9e88e802e2700cf272a0b6140ec # 0.20231123.829_ref.26ca90b",
		},
		{
			name: "valid 2",
			args: args{
				refstr: "devopsfaith/krakend:2.5.0",
			},
			want: "index.docker.io/devopsfaith/krakend@sha256:6a3c8e5e1a4948042bfb364ed6471e16b4a26d0afb6c3c01ebcb88b3fa551036 # 2.5.0",
		},
		{
			name: "invalid ref string",
			args: args{
				refstr: "ghcr.io/stacklok/minder/helm/minder!",
			},
			wantErr: true,
		},
		{
			name: "nonexistent container in nonexistent registry",
			args: args{
				refstr: "beeeeer.io/ipa/toppling-goliath/king-sue:1.0.0",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			r := New(&config.Config{})
			got, err := r.ParseSingleContainerImage(ctx, tt.args.refstr)
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

func TestReplacer_ParseSingleGitHubAction(t *testing.T) {
	t.Parallel()

	type args struct {
		action string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "actions/checkout with v4.1.1",
			args: args{
				action: "actions/checkout@v4.1.1",
			},
			want:    "actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1",
			wantErr: false,
		},
		{
			name: "actions/checkout with v3.6.0",
			args: args{
				action: "actions/checkout@v3.6.0",
			},
			want:    "actions/checkout@f43a0e5ff2bd294095638e18286ca9a3d1956744 # v3.6.0",
			wantErr: false,
		},
		{
			name: "actions/checkout with checksum returns checksum",
			args: args{
				action: "actions/checkout@1d96c772d19495a3b5c517cd2bc0cb401ea0529f",
			},
			want:    "1d96c772d19495a3b5c517cd2bc0cb401ea0529f",
			wantErr: false,
		},
		{
			name: "aquasecurity/trivy-action with 0.14.0",
			args: args{
				action: "aquasecurity/trivy-action@0.14.0",
			},
			want:    "2b6a709cf9c4025c5438138008beaddbb02086f0",
			wantErr: false,
		},
		{
			name: "aquasecurity/trivy-action with branch returns checksum",
			args: args{
				action: "aquasecurity/trivy-action@bump-trivy",
			},
			want:    "fb5e1b36be448e92ca98648c661bd7e9da1f1317",
			wantErr: false,
		},
		{
			name: "actions/checkout with invalid tag returns error",
			args: args{
				action: "actions/checkout@v4.1.1.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "actions/checkout with invalid action returns error",
			args: args{
				action: "invalid-action@v4.1.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "actions/checkout with empty action returns error",
			args: args{
				action: "@v4.1.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "actions/checkout with empty tag returns error",
			args: args{
				action: "actions/checkout",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "bufbuild/buf-setup-action with v1 is an array",
			args: args{
				action: "bufbuild/buf-setup-action@v1",
			},
			want: "480a0ee8a588045b52a847b48138c6f377a89519",
		},
		{
			name: "anchore/sbom-action/download-syft with a sub-action works",
			args: args{
				action: "anchore/sbom-action/download-syft@v0.14.3",
			},
			want: "78fc58e266e87a38d4194b2137a3d4e9bcaf7ca1",
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := New(&config.Config{}).WithGitHubClient(os.Getenv("GITHUB_TOKEN"))
			got, err := r.ParseSingleGitHubAction(context.Background(), tt.args.action)
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
