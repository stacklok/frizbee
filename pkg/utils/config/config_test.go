package config

import (
	"context"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestFromCommand(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		contextCfg   *Config
		platformFlag string
		expectedCfg  *Config
		expectError  bool
	}{
		{
			name:        "NoConfigInContext",
			contextCfg:  nil,
			expectError: true,
		},
		{
			name:        "WithConfigInContext",
			contextCfg:  &Config{Platform: "linux/arm64"},
			expectedCfg: &Config{Platform: "linux/arm64"},
		},
		{
			name:         "WithPlatformFlag",
			contextCfg:   &Config{Platform: "linux/amd64"},
			platformFlag: "windows/arm64",
			expectedCfg:  &Config{Platform: "windows/arm64"},
		},
	}

	for _, tt := range testCases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			cmd := &cobra.Command{}
			if tt.contextCfg != nil {
				ctx := context.WithValue(ctx, ContextConfigKey, tt.contextCfg)
				cmd.SetContext(ctx)
			} else {
				cmd.SetContext(ctx)
			}
			if tt.platformFlag != "" {
				cmd.Flags().String("platform", "", "platform")
				require.NoError(t, cmd.Flags().Set("platform", tt.platformFlag))
			}

			cfg, err := FromCommand(cmd)
			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedCfg, cfg)
			}
		})
	}
}

func TestParseConfigFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		fsContent      map[string]string
		fileName       string
		expectedResult *Config
		expectError    bool
	}{
		{
			name:           "FileNotFound",
			fileName:       "nonexistent.yaml",
			expectedResult: defaultConfig(),
		},
		{
			name:        "InvalidYaml",
			fileName:    "invalid.yaml",
			fsContent:   map[string]string{"invalid.yaml": "invalid yaml content"},
			expectError: true,
		},
		{
			name:     "DontIngoreBranches",
			fileName: "dont_ignore_branches.yaml",
			fsContent: map[string]string{
				"dont_ignore_branches.yaml": `
platform: linux/amd64
ghactions:
  exclude_branches:
`,
			},
			expectedResult: &Config{
				Platform: "linux/amd64",
				GHActions: GHActions{
					Filter: Filter{
						ExcludeBranches: []string{},
					},
				},
				Images: Images{
					ImageFilter: ImageFilter{
						ExcludeImages: []string{"scratch"},
						ExcludeTags:   []string{"latest"},
					},
				},
			},
		},
		{
			name:     "ValidYaml",
			fileName: "valid.yaml",
			fsContent: map[string]string{
				"valid.yaml": `
platform: linux/amd64
ghactions:
  exclude:
    - pattern1
    - pattern2
images:
  exclude_images:
    - notthisone
  exclude_tags:
    - notthistag	
`,
			},
			expectedResult: &Config{
				Platform: "linux/amd64",
				GHActions: GHActions{
					Filter: Filter{
						Exclude:         []string{"pattern1", "pattern2"},
						ExcludeBranches: []string{"*"},
					},
				},
				Images: Images{
					ImageFilter: ImageFilter{
						ExcludeImages: []string{"notthisone"},
						ExcludeTags:   []string{"notthistag"},
					},
				},
			},
		},
		{
			name:           "EmptyFile",
			fileName:       "empty.yaml",
			fsContent:      map[string]string{"empty.yaml": ""},
			expectedResult: defaultConfig(),
		},
	}

	for _, tt := range testCases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := memfs.New()
			for name, content := range tt.fsContent {
				f, _ := fs.Create(name)
				_, _ = f.Write([]byte(content))
				require.NoError(t, f.Close())
			}

			cfg, err := ParseConfigFileFromFS(fs, tt.fileName)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResult.Platform, cfg.Platform)
				if cfg.GHActions.Exclude != nil {
					require.Equal(t, tt.expectedResult.GHActions.Exclude, cfg.GHActions.Exclude)
				}
				if cfg.Images.ExcludeImages != nil {
					require.Equal(t, tt.expectedResult.Images.ExcludeImages, cfg.Images.ExcludeImages)
				}
				if cfg.Images.ExcludeTags != nil {
					require.Equal(t, tt.expectedResult.Images.ExcludeTags, cfg.Images.ExcludeTags)
				}
				if cfg.GHActions.ExcludeBranches != nil {
					require.Equal(t, tt.expectedResult.GHActions.ExcludeBranches, cfg.GHActions.ExcludeBranches)
				}
			}
		})
	}
}
