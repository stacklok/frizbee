package image

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

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
	cfg := config.Config{GHActions: config.GHActions{Filter: config.Filter{Exclude: []string{"scratch"}}}}

	tests := []struct {
		name        string
		matchedLine string
	}{
		{"Replace excluded path", "FROM scratch"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := parser.Replace(ctx, tt.matchedLine, nil, cfg)
			require.Error(t, err, "Should return error for excluded path")
			require.Contains(t, err.Error(), "reference skipped", "Error should indicate reference skipped")
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

func TestShouldExclude(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ref  string
		want bool
	}{
		{"Exclude scratch", "scratch", true},
		{"Do not exclude ubuntu", "ubuntu", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := shouldExclude(tt.ref)
			require.Equal(t, tt.want, got, "shouldExclude should return the correct exclusion status")
		})
	}
}
