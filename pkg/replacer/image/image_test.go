package image

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stacklok/frizbee/pkg/interfaces"
	"github.com/stacklok/frizbee/pkg/utils/config"
	"github.com/stacklok/frizbee/pkg/utils/store"
)

func TestNewParser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{"New parser initialization"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parser := New()
			require.NotNil(t, parser, "Parser should not be nil")
			require.Equal(t, ContainerImageRegex, parser.regex, "Default regex should be ContainerImageRegex")
			require.NotNil(t, parser.cache, "Cache should be initialized")
		})
	}
}

func TestSetCache(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		cache store.RefCacher
	}{
		{"Set cache for parser", store.NewRefCacher()},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parser := New()
			parser.SetCache(tt.cache)
			require.Equal(t, tt.cache, parser.cache, "Cache should be set correctly")
		})
	}
}

func TestSetAndGetRegex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		newRegex string
	}{
		{"Set and get new regex", `new-regex`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parser := New()
			parser.SetRegex(tt.newRegex)
			require.Equal(t, tt.newRegex, parser.GetRegex(), "Regex should be set and retrieved correctly")
		})
	}
}

func TestReplaceExcludedPath(t *testing.T) {
	t.Parallel()

	parser := New()
	ctx := context.Background()
	cfg := config.Config{
		Images: config.Images{
			ImageFilter: config.ImageFilter{
				ExcludeImages: []string{"scratch"},
				ExcludeTags:   []string{"latest"},
			},
		},
	}

	tests := []struct {
		name        string
		matchedLine string
		expected    error
	}{
		{
			"Do not replace scratch FROM image",
			"FROM scratch",
			interfaces.ErrReferenceSkipped,
		},
		{
			"Do not replace ubuntu:latest",
			"FROM ubuntu:latest",
			interfaces.ErrReferenceSkipped,
		},
		{
			"Do not replace ubuntu:latest with AS",
			"FROM ubuntu:latest AS builder",
			interfaces.ErrReferenceSkipped,
		},
		{
			"Do not replace ubuntu without a tag",
			"FROM ubuntu",
			interfaces.ErrReferenceSkipped,
		},
		{
			"Do not replace ubuntu without a tag with a stage",
			"FROM ubuntu AS builder",
			interfaces.ErrReferenceSkipped,
		},
		{
			"Replace ubuntu:22.04",
			"FROM ubuntu:22.04",
			nil,
		},
		{
			"Replace ubuntu:22.04 with AS",
			"FROM ubuntu:22.04 AS builder",
			nil,
		},
		{
			"Replace ubuntu:22.04 with AS",
			"FROM --platform=linux/amd64 ubuntu:22.04 AS builder",
			nil,
		},
		{
			"Replace with repo reference and tag",
			"FROM ghcr.io/stacklok/minder/helm/minder:0.20231123.829_ref.26ca90b",
			nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := parser.Replace(ctx, tt.matchedLine, nil, cfg)
			if tt.expected == nil {
				require.NoError(t, err, "Should not return error for excluded path")
			} else {
				require.Error(t, err, "Should return error for excluded path")
				require.ErrorIs(t, err, tt.expected, "Unexpected error")
			}
		})
	}
}

func TestConvertToEntityRef(t *testing.T) {
	t.Parallel()

	parser := New()

	tests := []struct {
		name      string
		reference string
		wantErr   bool
	}{
		{"Valid container reference with tag", "ghcr.io/stacklok/minder/helm/minder:0.20231123.829_ref.26ca90b", false},
		{"Valid container reference with digest", "ghcr.io/stacklok/minder/helm/minder@sha256:a29f8a8d28f0af7f70a4b3dd3e33c8c8cc5cf9e88e802e2700cf272a0b6140ec", false},
		{"Invalid reference format", "invalid:reference:format", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ref, err := parser.ConvertToEntityRef(tt.reference)
			if tt.wantErr {
				require.Error(t, err, "Expected error but got none")
			} else {
				require.NoError(t, err, "Expected no error but got %v", err)
				require.NotNil(t, ref, "EntityRef should not be nil")
			}
		})
	}
}

func TestGetImageDigestFromRef(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name    string
		refstr  string
		want    string
		wantErr bool
	}{
		{
			name:   "Valid image reference 1",
			refstr: "ghcr.io/stacklok/minder/helm/minder:0.20231123.829_ref.26ca90b",
			want:   "sha256:a29f8a8d28f0af7f70a4b3dd3e33c8c8cc5cf9e88e802e2700cf272a0b6140ec",
		},
		{
			name:   "Valid image reference 2",
			refstr: "devopsfaith/krakend:2.5.0",
			want:   "sha256:6a3c8e5e1a4948042bfb364ed6471e16b4a26d0afb6c3c01ebcb88b3fa551036",
		},
		{
			name:    "Invalid ref string",
			refstr:  "ghcr.io/stacklok/minder/helm/minder!",
			wantErr: true,
		},
		{
			name:    "Nonexistent container in nonexistent registry",
			refstr:  "beeeeer.io/ipa/toppling-goliath/king-sue:1.0.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := GetImageDigestFromRef(ctx, tt.refstr, "", nil)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got.Ref)
		})
	}
}

func TestShouldSkipImage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ref  string
		skip bool
	}{
		// skip cases
		{"Skip scratch", "scratch", true},
		{"Skip ubuntu without a tag", "ubuntu", true},
		{"Skip ubuntu:latest", "ubuntu:latest", true},
		// keep cases
		{"Do not skip ubuntu:22.04", "ubuntu:22.04", false},
		{"Do not skip with repo reference and tag", "myrepo/myimage:1.2.3", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := &config.Config{
				Images: config.Images{
					ImageFilter: config.ImageFilter{
						ExcludeImages: []string{"scratch"},
						ExcludeTags:   []string{"latest"},
					},
				},
			}

			got := shouldSkipImageRef(config, tt.ref)
			require.Equal(t, tt.skip, got, "shouldSkipImageRef should return the correct exclusion status")
		})
	}
}
