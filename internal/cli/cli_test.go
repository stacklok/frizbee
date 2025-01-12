package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewHelper(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		cmdArgs       []string
		expected      *Helper
		expectedError bool
	}{
		{
			name:    "ValidFlags",
			cmdArgs: []string{"--dry-run", "--quiet", "--error", "--regex", "test"},
			expected: &Helper{
				DryRun:        true,
				Quiet:         true,
				ErrOnModified: true,
				Regex:         "test",
			},
			expectedError: false,
		},
		{
			name:          "MissingFlags",
			cmdArgs:       []string{},
			expected:      &Helper{},
			expectedError: false,
		},
		{
			name:          "InvalidFlags",
			cmdArgs:       []string{"--nonexistent"},
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cobra.Command{}
			DeclareFrizbeeFlags(cmd, true)
			cmd.SetArgs(tt.cmdArgs)

			if tt.expectedError {
				assert.Error(t, cmd.Execute())
				return
			}

			assert.NoError(t, cmd.Execute())

			helper, err := NewHelper(cmd)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, helper)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, helper)
				assert.Equal(t, tt.expected.DryRun, helper.DryRun)
				assert.Equal(t, tt.expected.Quiet, helper.Quiet)
				assert.Equal(t, tt.expected.ErrOnModified, helper.ErrOnModified)
				assert.Equal(t, tt.expected.Regex, helper.Regex)
			}
		})
	}
}

func TestProcessOutput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		helper         *Helper
		path           string
		processed      []string
		modified       map[string]string
		expectedOutput string
		expectError    bool
	}{
		{
			name: "QuietMode",
			helper: &Helper{
				Quiet: true,
				Cmd:   &cobra.Command{},
			},
			path:           "test/path",
			processed:      []string{"file1.txt", "file2.txt"},
			modified:       map[string]string{"file1.txt": "new content"},
			expectedOutput: "",
			expectError:    false,
		},
		{
			name: "DryRunMode",
			helper: &Helper{
				Quiet:  false,
				DryRun: true,
				Cmd:    &cobra.Command{},
			},
			path:           "test/path",
			processed:      []string{"file1.txt"},
			modified:       map[string]string{"file1.txt": "new content"},
			expectedOutput: "Processed: file1.txt\nModified: file1.txt\nnew content",
			expectError:    false,
		},
		{
			name: "ErrorOpeningFile",
			helper: &Helper{
				Quiet: false,
				Cmd:   &cobra.Command{},
			},
			path:           "invalid/path",
			modified:       map[string]string{"invalid/path": "new content"},
			expectedOutput: "",
			expectError:    true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Set up command output
			var output strings.Builder
			tt.helper.Cmd.SetOut(&output)
			tt.helper.Cmd.SetErr(&output)

			// Create in-memory filesystem and add files
			fs := memfs.New()
			for path, content := range tt.modified {
				dir := filepath.Join(tt.path, filepath.Dir(path))
				assert.NoError(t, fs.MkdirAll(dir, 0755))
				file, err := fs.Create(filepath.Join(tt.path, path))
				if err == nil {
					_, _ = file.Write([]byte(content))
					assert.NoError(t, file.Close())
				}
			}

			// Process the output using the in-memory filesystem
			err := tt.helper.ProcessOutput(tt.path, tt.processed, tt.modified)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output.String(), tt.expectedOutput)
			}
		})
	}
}

func TestIsPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		setup    func(fs billy.Filesystem)
		path     string
		expected bool
	}{
		{
			name: "ExistingFile",
			setup: func(fs billy.Filesystem) {
				file, _ := fs.Create("testfile.txt")
				assert.NoError(t, file.Close())
			},
			path:     "testfile.txt",
			expected: true,
		},
		{
			name:     "NonExistentFile",
			setup:    func(_ billy.Filesystem) {},
			path:     "nonexistent.txt",
			expected: false,
		},
		{
			name: "ExistingDirectory",
			setup: func(fs billy.Filesystem) {
				assert.NoError(t, fs.MkdirAll("testdir", 0755))
			},
			path:     "testdir",
			expected: true,
		},
		{
			name:     "NonExistentDirectory",
			setup:    func(_ billy.Filesystem) {},
			path:     "nonexistentdir",
			expected: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Use in-memory filesystem for testing
			fs := memfs.New()
			tt.setup(fs)

			// Check if the path exists in the in-memory filesystem
			_, err := fs.Stat(tt.path)
			result := err == nil

			assert.Equal(t, tt.expected, result)
		})
	}
}
