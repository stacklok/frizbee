package actions

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stacklok/frizbee/pkg/utils/config"
	"github.com/stacklok/frizbee/pkg/utils/ghrest"
	"github.com/stacklok/frizbee/pkg/utils/store"
)

func TestNewParser(t *testing.T) {
	t.Parallel()

	parser := New()
	require.NotNil(t, parser, "Parser should not be nil")
	require.Equal(t, GitHubActionsRegex, parser.regex, "Default regex should be GitHubActionsRegex")
	require.NotNil(t, parser.cache, "Cache should be initialized")
}

func TestSetCache(t *testing.T) {
	t.Parallel()

	parser := New()
	cache := store.NewRefCacher()
	parser.SetCache(cache)
	require.Equal(t, cache, parser.cache, "Cache should be set correctly")
}

func TestSetAndGetRegex(t *testing.T) {
	t.Parallel()

	parser := New()
	tests := []struct {
		name     string
		newRegex string
	}{
		{
			name:     "Set and get new regex",
			newRegex: `new-regex`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser.SetRegex(tt.newRegex)
			require.Equal(t, tt.newRegex, parser.GetRegex(), "Regex should be set and retrieved correctly")
		})
	}
}

func TestReplaceLocalPath(t *testing.T) {
	t.Parallel()

	parser := New()
	ctx := context.Background()
	cfg := config.Config{}
	restIf := &ghrest.Client{}

	tests := []struct {
		name        string
		matchedLine string
	}{
		{
			name:        "Replace local path",
			matchedLine: "./local/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := parser.Replace(ctx, tt.matchedLine, restIf, cfg)
			require.Error(t, err, "Should return error for local path")
			require.Contains(t, err.Error(), "reference skipped", "Error should indicate reference skipped")
		})
	}
}

func TestReplaceExcludedPath(t *testing.T) {
	t.Parallel()

	parser := New()
	ctx := context.Background()
	cfg := config.Config{GHActions: config.GHActions{Filter: config.Filter{Exclude: []string{"actions/checkout"}}}}
	restIf := &ghrest.Client{}

	tests := []struct {
		name        string
		matchedLine string
	}{
		{
			name:        "Replace excluded path",
			matchedLine: "uses: actions/checkout@v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := parser.Replace(ctx, tt.matchedLine, restIf, cfg)
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
		{"Valid action reference", "uses: actions/checkout@v2", false},
		{"Valid docker reference", "docker://mydocker/image:tag", false},
		{"Invalid reference format", "invalid-reference", true},
	}

	for _, tt := range tests {
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

func TestIsLocal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Local path with ./", "./local/path", true},
		{"Local path with ../", "../local/path", true},
		{"Non-local path", "non/local/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, isLocal(tt.input), "IsLocal should return correct value")
		})
	}
}

func TestShouldExclude(t *testing.T) {
	t.Parallel()

	cfg := &config.GHActions{Filter: config.Filter{Exclude: []string{"actions/checkout", "actions/setup"}}}

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Excluded path", "actions/checkout", true},
		{"Non-excluded path", "actions/unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, shouldExclude(cfg, tt.input), "ShouldExclude should return correct value")
		})
	}
}

func TestParseActionReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantAction string
		wantRef    string
		wantErr    bool
	}{
		{"Valid action reference", "actions/checkout@v2", "actions/checkout", "v2", false},
		{"Invalid reference format", "invalid-reference", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			action, ref, err := ParseActionReference(tt.input)
			if tt.wantErr {
				require.Error(t, err, "Expected error but got none")
			} else {
				require.NoError(t, err, "Expected no error but got %v", err)
				require.Equal(t, tt.wantAction, action, "Action should be parsed correctly")
				require.Equal(t, tt.wantRef, ref, "Reference should be parsed correctly")
			}
		})
	}
}

func TestParseActionReference_WithQuotes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          string
		expectedAction string
		expectedRef    string
		expectError    bool
	}{
		{
			name:           "quoted action reference",
			input:          "'google-github-actions/auth@6fc4af4b145ae7821d527454aa9bd537d1f2dc5f'",
			expectedAction: "google-github-actions/auth",
			expectedRef:    "6fc4af4b145ae7821d527454aa9bd537d1f2dc5f",
			expectError:    false,
		},
		{
			name:           "double quoted action reference",
			input:          "\"google-github-actions/auth@6fc4af4b145ae7821d527454aa9bd537d1f2dc5f\"",
			expectedAction: "google-github-actions/auth",
			expectedRef:    "6fc4af4b145ae7821d527454aa9bd537d1f2dc5f",
			expectError:    false,
		},
		{
			name:           "unquoted action reference",
			input:          "google-github-actions/auth@6fc4af4b145ae7821d527454aa9bd537d1f2dc5f",
			expectedAction: "google-github-actions/auth",
			expectedRef:    "6fc4af4b145ae7821d527454aa9bd537d1f2dc5f",
			expectError:    false,
		},
		{
			name:        "invalid action reference",
			input:       "invalid-action-reference",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			action, ref, err := ParseActionReference(tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedAction, action)
			require.Equal(t, tc.expectedRef, ref)
		})
	}
}

func TestConvertToEntityRef_WithQuotes(t *testing.T) {
	t.Parallel()

	parser := New()

	testCases := []struct {
		name         string
		input        string
		expectedName string
		expectedRef  string
		expectedType string
		expectError  bool
	}{
		{
			name:         "quoted action reference with uses prefix",
			input:        "uses: 'google-github-actions/auth@6fc4af4b145ae7821d527454aa9bd537d1f2dc5f'",
			expectedName: "google-github-actions/auth",
			expectedRef:  "6fc4af4b145ae7821d527454aa9bd537d1f2dc5f",
			expectedType: "action",
			expectError:  false,
		},
		{
			name:         "double quoted action reference with uses prefix",
			input:        "uses: \"google-github-actions/auth@6fc4af4b145ae7821d527454aa9bd537d1f2dc5f\"",
			expectedName: "google-github-actions/auth",
			expectedRef:  "6fc4af4b145ae7821d527454aa9bd537d1f2dc5f",
			expectedType: "action",
			expectError:  false,
		},
		{
			name:         "unquoted action reference with uses prefix",
			input:        "uses: google-github-actions/auth@6fc4af4b145ae7821d527454aa9bd537d1f2dc5f",
			expectedName: "google-github-actions/auth",
			expectedRef:  "6fc4af4b145ae7821d527454aa9bd537d1f2dc5f",
			expectedType: "action",
			expectError:  false,
		},
		{
			name:        "invalid action reference",
			input:       "uses: invalid-action-reference",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := parser.ConvertToEntityRef(tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedName, result.Name)
			require.Equal(t, tc.expectedRef, result.Ref)
			require.Equal(t, tc.expectedType, result.Type)
		})
	}
}

func TestGetChecksum(t *testing.T) {
	t.Parallel()

	tok := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()
	ghcli := ghrest.NewClient(tok)

	tests := []struct {
		name    string
		args    struct{ action, ref string }
		want    string
		wantErr bool
	}{
		{
			name:    "actions/checkout with v4.1.1",
			args:    struct{ action, ref string }{action: "actions/checkout", ref: "v4.1.1"},
			want:    "b4ffde65f46336ab88eb53be808477a3936bae11",
			wantErr: false,
		},
		{
			name:    "actions/checkout with v3.6.0",
			args:    struct{ action, ref string }{action: "actions/checkout", ref: "v3.6.0"},
			want:    "f43a0e5ff2bd294095638e18286ca9a3d1956744",
			wantErr: false,
		},
		{
			name:    "actions/checkout with checksum returns checksum",
			args:    struct{ action, ref string }{action: "actions/checkout", ref: "1d96c772d19495a3b5c517cd2bc0cb401ea0529f"},
			want:    "1d96c772d19495a3b5c517cd2bc0cb401ea0529f",
			wantErr: false,
		},
		{
			name:    "aquasecurity/trivy-action with 0.14.0",
			args:    struct{ action, ref string }{action: "aquasecurity/trivy-action", ref: "0.14.0"},
			want:    "2b6a709cf9c4025c5438138008beaddbb02086f0",
			wantErr: false,
		},
		{
			name:    "aquasecurity/trivy-action with branch returns checksum",
			args:    struct{ action, ref string }{action: "aquasecurity/trivy-action", ref: "bump-trivy"},
			want:    "fb5e1b36be448e92ca98648c661bd7e9da1f1317",
			wantErr: false,
		},
		{
			name:    "actions/checkout with invalid tag returns error",
			args:    struct{ action, ref string }{action: "actions/checkout", ref: "v4.1.1.1"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "actions/checkout with invalid action returns error",
			args:    struct{ action, ref string }{action: "invalid-action", ref: "v4.1.1"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "actions/checkout with empty action returns error",
			args:    struct{ action, ref string }{action: "", ref: "v4.1.1"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "actions/checkout with empty tag returns error",
			args:    struct{ action, ref string }{action: "actions/checkout", ref: ""},
			want:    "",
			wantErr: true,
		},
		{
			name:    "actions/setup-node with v1 is an array",
			args:    struct{ action, ref string }{action: "actions/setup-node", ref: "v1"},
			want:    "f1f314fca9dfce2769ece7d933488f076716723e",
			wantErr: false,
		},
		{
			name:    "anchore/sbom-action/download-syft with a sub-action works",
			args:    struct{ action, ref string }{action: "anchore/sbom-action/download-syft", ref: "v0.14.3"},
			want:    "78fc58e266e87a38d4194b2137a3d4e9bcaf7ca1",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := GetChecksum(ctx, config.GHActions{}, ghcli, tt.args.action, tt.args.ref)
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
