package traverse

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/assert"
)

func TestYamlDockerfiles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		fsContent   map[string]string
		baseDir     string
		expected    []string
		expectError bool
	}{
		{
			name: "NoYAMLOrDockerfile",
			fsContent: map[string]string{
				"base/file.txt": "content",
			},
			baseDir:     "base",
			expected:    []string{},
			expectError: false,
		},
		{
			name: "WithYAMLFiles",
			fsContent: map[string]string{
				"base/file.yml":         "content",
				"base/file.yaml":        "content",
				"base/not_included.txt": "content",
			},
			baseDir: "base",
			expected: []string{
				"base/file.yml",
				"base/file.yaml",
			},
			expectError: false,
		},
		{
			name: "WithDockerfiles",
			fsContent: map[string]string{
				"base/Dockerfile":        "content",
				"base/nested/dockerfile": "content",
				"base/not_included.txt":  "content",
			},
			baseDir: "base",
			expected: []string{
				"base/Dockerfile",
				"base/nested/dockerfile",
			},
			expectError: false,
		},
		{
			name: "MixedFiles",
			fsContent: map[string]string{
				"base/file.yml":          "content",
				"base/Dockerfile":        "content",
				"base/nested/file.yaml":  "content",
				"base/nested/dockerfile": "content",
				"base/not_included.txt":  "content",
			},
			baseDir: "base",
			expected: []string{
				"base/file.yml",
				"base/Dockerfile",
				"base/nested/file.yaml",
				"base/nested/dockerfile",
			},
			expectError: false,
		},
		{
			name: "ErrorInProcessingFile",
			fsContent: map[string]string{
				"base/file.yml": "content",
			},
			baseDir:     "base",
			expectError: true,
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
				assert.NoError(t, f.Close())
			}

			var processedFiles []string
			err := YamlDockerfiles(fs, tt.baseDir, func(path string) error {
				if tt.expectError {
					return errors.New("error in processing file")
				}
				processedFiles = append(processedFiles, path)
				return nil
			})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expected, processedFiles)
			}
		})
	}
}

func TestTraverse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		fsContent   map[string]string
		baseDir     string
		expected    []string
		expectError bool
	}{
		{
			name: "TraverseFiles",
			fsContent: map[string]string{
				"base/file1.txt":   "content",
				"base/file2.txt":   "content",
				"base/nested/file": "content",
			},
			baseDir: "base",
			expected: []string{
				"base",
				"base/file1.txt",
				"base/file2.txt",
				"base/nested",
				"base/nested/file",
			},
			expectError: false,
		},
		{
			name: "TraverseWithError",
			fsContent: map[string]string{
				"base/file.txt": "content",
			},
			baseDir:     "base",
			expectError: true,
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
				assert.NoError(t, f.Close())
			}

			var processedFiles []string
			err := Traverse(fs, tt.baseDir, func(path string, _ os.FileInfo) error {
				if tt.expectError {
					return errors.New("error in traversing file")
				}
				processedFiles = append(processedFiles, path)
				return nil
			})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expected, processedFiles)
			}
		})
	}
}

func TestIsYAMLOrDockerfile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		fileName string
		isDir    bool
		expected bool
	}{
		{
			name:     "YAMLFile",
			fileName: "config.yaml",
			isDir:    false,
			expected: true,
		},
		{
			name:     "YMLFile",
			fileName: "config.yml",
			isDir:    false,
			expected: true,
		},
		{
			name:     "Dockerfile",
			fileName: "Dockerfile",
			isDir:    false,
			expected: true,
		},
		{
			name:     "dockerfile",
			fileName: "dockerfile",
			isDir:    false,
			expected: true,
		},
		{
			name:     "NonYAMLOrDockerfile",
			fileName: "config.txt",
			isDir:    false,
			expected: false,
		},
		{
			name:     "Directory",
			fileName: "config",
			isDir:    true,
			expected: false,
		},
	}

	for _, tt := range testCases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info := &fileInfoMock{
				name: tt.fileName,
				dir:  tt.isDir,
			}

			result := isYAMLOrDockerfile(info)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// fileInfoMock is a mock implementation of os.FileInfo for testing.
type fileInfoMock struct {
	name string
	dir  bool
}

func (f *fileInfoMock) Name() string       { return f.name }
func (_ *fileInfoMock) Size() int64        { return 0 }
func (_ *fileInfoMock) Mode() os.FileMode  { return 0 }
func (_ *fileInfoMock) ModTime() time.Time { return time.Time{} }
func (f *fileInfoMock) IsDir() bool        { return f.dir }
func (_ *fileInfoMock) Sys() interface{}   { return nil }
