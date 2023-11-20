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
	"testing"

	"github.com/google/go-github/v56/github"
	"github.com/stretchr/testify/require"

	"github.com/stacklok/frisbee/pkg/ghactions"
)

func TestParseActionReference(t *testing.T) {
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
			name: "actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11",
			args: args{
				input: "actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11",
			},
			returns: returns{
				action:    "actions/checkout",
				reference: "b4ffde65f46336ab88eb53be808477a3936bae11",
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
				ref:    "b4ffde65f46336ab88eb53be808477a3936bae11",
			},
			want:    "b4ffde65f46336ab88eb53be808477a3936bae11",
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
				action: "actions/checkout/invalid",
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
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ghcli := github.NewClient(nil)
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
