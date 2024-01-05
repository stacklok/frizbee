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

	"github.com/google/go-containerregistry/pkg/name"
	gocrv1 "github.com/google/go-containerregistry/pkg/v1"
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

func TestGetDigestFromRef(t *testing.T) {
	t.Parallel()

	type args struct {
		refstr string
	}
	type flags struct {
		platform *gocrv1.Platform
	}
	tests := []struct {
		name    string
		args    args
		flags   flags
		want    string
		wantErr bool
	}{
		{
			name: "valid image without platform",
			args: args{
				refstr: "registry.k8s.io/kube-apiserver:v1.20.0",
			},
			want: "sha256:8b8125d7a6e4225b08f04f65ca947b27d0cc86380bf09fab890cc80408230114",
		},
		{
			name: "valid arm64 image",
			args: args{
				refstr: "registry.k8s.io/kube-apiserver:v1.20.0",
			},
			flags: flags{
				platform: &gocrv1.Platform{Architecture: "arm64", OS: "linux"},
			},
			want: "sha256:36464375c04fad5847fa5ab371acb9786e49e75a4261dbdfae739593e310c72f",
		},
		{
			name: "valid amd64 image",
			args: args{
				refstr: "registry.k8s.io/kube-apiserver:v1.20.0",
			},
			flags: flags{
				platform: &gocrv1.Platform{Architecture: "amd64", OS: "linux"},
			},
			want: "sha256:8033693d4421e41bd91380ce3c6b1a20fbaf762e3c4d64f79bbb3e30a2fb4310",
		},
		{
			name: "Invalid architecture",
			args: args{
				refstr: "registry.k8s.io/kube-apiserver:v1.20.0",
			},
			flags: flags{
				platform: &gocrv1.Platform{Architecture: "foo", OS: "linux"},
			},
			wantErr: true,
		},
		{
			name: "Invalid OS",
			args: args{
				refstr: "registry.k8s.io/kube-apiserver:v1.20.0",
			},
			flags: flags{
				platform: &gocrv1.Platform{Architecture: "amd64", OS: "foo"},
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
			reference, err := name.ParseReference(tt.args.refstr)
			if err != nil {
				t.Fatalf("failed to parse reference: %v", err)
			}
			SetPlatform(tt.flags.platform)
			got, err := GetDigestFromRef(ctx, reference)
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
