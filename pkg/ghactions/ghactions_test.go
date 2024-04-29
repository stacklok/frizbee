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

package ghactions_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/stacklok/frizbee/internal/ghrest"
	"github.com/stacklok/frizbee/pkg/config"
	"github.com/stacklok/frizbee/pkg/ghactions"
)

func TestParseActionReference(t *testing.T) {
	t.Parallel()

	type args struct {
		input string
	}
	type returns struct {
		action    string
		reference string
	}
	tests := []struct {
		name    string
		args    args
		returns returns
		wantErr bool
	}{
		{
			name: "actions/checkout@v4.1.1",
			args: args{
				input: "actions/checkout@v4.1.1",
			},
			returns: returns{
				action:    "actions/checkout",
				reference: "v4.1.1",
			},
			wantErr: false,
		},
		{
			name: "actions/checkout@v3.6.0",
			args: args{
				input: "actions/checkout@v3.6.0",
			},
			returns: returns{
				action:    "actions/checkout",
				reference: "v3.6.0",
			},
			wantErr: false,
		},
		{
			name: "actions/checkout@1d96c772d19495a3b5c517cd2bc0cb401ea0529f",
			args: args{
				input: "actions/checkout@1d96c772d19495a3b5c517cd2bc0cb401ea0529f",
			},
			returns: returns{
				action:    "actions/checkout",
				reference: "1d96c772d19495a3b5c517cd2bc0cb401ea0529f",
			},
			wantErr: false,
		},
		{
			name: "actions/checkout-invalid",
			args: args{
				input: "actions/checkout-invalid",
			},
			returns: returns{
				action:    "",
				reference: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotact, gotref, err := ghactions.ParseActionReference(tt.args.input)
			if tt.wantErr {
				require.Error(t, err, "Wanted error, got none")
				return
			}
			require.NoError(t, err, "Wanted no error, got %v", err)
			require.Equal(t, tt.returns.action, gotact, "Wanted %v, got %v", tt.returns.action, gotact)
			require.Equal(t, tt.returns.reference, gotref, "Wanted %v, got %v", tt.returns.reference, gotref)
		})
	}
}

func TestGetChecksum(t *testing.T) {
	t.Parallel()

	tok := os.Getenv("GITHUB_TOKEN")

	type args struct {
		action string
		ref    string
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
				action: "actions/checkout",
				ref:    "v4.1.1",
			},
			want:    "b4ffde65f46336ab88eb53be808477a3936bae11",
			wantErr: false,
		},
		{
			name: "actions/checkout with v3.6.0",
			args: args{
				action: "actions/checkout",
				ref:    "v3.6.0",
			},
			want:    "f43a0e5ff2bd294095638e18286ca9a3d1956744",
			wantErr: false,
		},
		{
			name: "actions/checkout with checksum returns checksum",
			args: args{
				action: "actions/checkout",
				ref:    "1d96c772d19495a3b5c517cd2bc0cb401ea0529f",
			},
			want:    "1d96c772d19495a3b5c517cd2bc0cb401ea0529f",
			wantErr: false,
		},
		{
			name: "aquasecurity/trivy-action with 0.14.0",
			args: args{
				action: "aquasecurity/trivy-action",
				ref:    "0.14.0",
			},
			want:    "2b6a709cf9c4025c5438138008beaddbb02086f0",
			wantErr: false,
		},
		{
			name: "aquasecurity/trivy-action with branch returns checksum",
			args: args{
				action: "aquasecurity/trivy-action",
				ref:    "bump-trivy",
			},
			want:    "fb5e1b36be448e92ca98648c661bd7e9da1f1317",
			wantErr: false,
		},
		{
			name: "actions/checkout with invalid tag returns error",
			args: args{
				action: "actions/checkout",
				ref:    "v4.1.1.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "actions/checkout with invalid action returns error",
			args: args{
				action: "invalid-action",
				ref:    "v4.1.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "actions/checkout with empty action returns error",
			args: args{
				action: "",
				ref:    "v4.1.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "actions/checkout with empty tag returns error",
			args: args{
				action: "actions/checkout",
				ref:    "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "bufbuild/buf-setup-action with v1 is an array",
			args: args{
				action: "bufbuild/buf-setup-action",
				ref:    "v1",
			},
			want: "480a0ee8a588045b52a847b48138c6f377a89519",
		},
		{
			name: "anchore/sbom-action/download-syft with a sub-action works",
			args: args{
				action: "anchore/sbom-action/download-syft",
				ref:    "v0.14.3",
			},
			want: "78fc58e266e87a38d4194b2137a3d4e9bcaf7ca1",
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ghcli := ghrest.NewGhRest(tok)
			got, err := ghactions.GetChecksum(context.Background(), ghcli, tt.args.action, tt.args.ref)
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

const (
	workflowYAML = `
name: CI
on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: checkout
        uses: actions/checkout@v4
      - name: setup go
        uses: actions/setup-go@v4
`
)

func TestModifyReferencesInYAML(t *testing.T) {
	t.Parallel()

	tok := os.Getenv("GITHUB_TOKEN")

	tests := []struct {
		name           string
		mustContain    []string
		mustNotContain []string
		wantErr        bool
		cfg            *config.GHActions
	}{
		{
			name:    "modify all",
			wantErr: false,
			mustContain: []string{
				"              uses: actions/checkout@0ad4b8fadaa221de15dcec353f45205ec38ea70b # v4",
				"              uses: actions/setup-go@93397bea11091df50f3d7e59dc26a7711a8bcfbe # v4",
			},
			mustNotContain: []string{
				"              uses: actions/checkout@v4",
				"              uses: actions/setup-go@v4",
			},
			cfg: &config.GHActions{
				Filter: config.Filter{
					Exclude: []string{},
				},
			},
		},
		{
			name:    "exclude full uses",
			wantErr: false,
			mustContain: []string{
				"              uses: actions/checkout@v4",
				"              uses: actions/setup-go@93397bea11091df50f3d7e59dc26a7711a8bcfbe # v4",
			},
			mustNotContain: []string{
				"              uses: actions/checkout@0ad4b8fadaa221de15dcec353f45205ec38ea70b # v4",
				"              uses: actions/setup-go@v4",
			},
			cfg: &config.GHActions{
				Filter: config.Filter{
					Exclude: []string{
						"actions/checkout@v4",
					},
				},
			},
		},
		{
			name:    "exclude just the action name",
			wantErr: false,
			mustContain: []string{
				"              uses: actions/checkout@v4",
				"              uses: actions/setup-go@93397bea11091df50f3d7e59dc26a7711a8bcfbe # v4",
			},
			mustNotContain: []string{
				"              uses: actions/checkout@0ad4b8fadaa221de15dcec353f45205ec38ea70b # v4",
				"              uses: actions/setup-go@v4",
			},
			cfg: &config.GHActions{
				Filter: config.Filter{
					Exclude: []string{
						"actions/checkout",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ghcli := ghrest.NewGhRest(tok)

			var root yaml.Node
			err := yaml.Unmarshal([]byte(workflowYAML), &root)
			require.NoError(t, err, "Error unmarshalling YAML, got %v", err)

			got, err := ghactions.ModifyReferencesInYAML(context.Background(), ghcli, &root, tt.cfg)
			if tt.wantErr {
				require.Error(t, err, "Wanted error, got none")
				require.Empty(t, got, "Wanted empty string, got %v", got)
				return
			}
			require.NoError(t, err, "Wanted no error, got %v", err)

			out, err := yaml.Marshal(&root)
			require.NoError(t, err, "Error marhsalling YAML, got %v", err)
			stringSlice := strings.Split(string(out), "\n")

			require.Subset(t, stringSlice, tt.mustContain, "Expected %v to not appear in %v", tt.mustContain, stringSlice)
			require.NotSubset(t, stringSlice, tt.mustNotContain, "Expected %v to not appear in %v", tt.mustNotContain, stringSlice)
		})
	}
}
