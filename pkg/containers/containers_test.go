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

// Package containers provides functions to replace tags for checksums
package containers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDigest(t *testing.T) {
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
			want: "sha256:a29f8a8d28f0af7f70a4b3dd3e33c8c8cc5cf9e88e802e2700cf272a0b6140ec",
		},
		{
			name: "valid 2",
			args: args{
				refstr: "devopsfaith/krakend:2.5.0",
			},
			want: "sha256:6a3c8e5e1a4948042bfb364ed6471e16b4a26d0afb6c3c01ebcb88b3fa551036",
		},
		{
			name: "invalid ref string",
			args: args{
				refstr: "ghcr.io/stacklok/minder/helm/minder!",
			},
			wantErr: true,
		},
		{
			name: "unexistent container in unexistent registry",
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
			got, err := GetDigest(ctx, tt.args.refstr)
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
