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

package ghactions

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/go-git/go-billy/v5"
	billyutil "github.com/go-git/go-billy/v5/util"
	"gopkg.in/yaml.v3"
)

// TraverseFunc is a function that gets called with each file in a GitHub Actions workflow
// directory. It receives the path to the file and the parsed workflow.
type TraverseFunc func(path string, wflow *yaml.Node) error

// TraverseGitHubActionWorkflows traverses the GitHub Actions workflows in the given directory
// and calls the given function with each workflow.
func TraverseGitHubActionWorkflows(bfs billy.Filesystem, base string, fun TraverseFunc) error {
	return billyutil.Walk(bfs, base, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("failed to walk path %s: %v\n", path, err)
			return nil
		}

		if shouldSkipFile(info) {
			fmt.Printf("skipping file %s\n", path)
			return nil
		}

		f, err := bfs.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer f.Close()

		dec := yaml.NewDecoder(f)

		var wflow yaml.Node
		if err := dec.Decode(&wflow); err != nil {
			return fmt.Errorf("failed to decode file %s: %w", path, err)
		}

		if err := fun(path, &wflow); err != nil {
			return fmt.Errorf("failed to process file %s: %w", path, err)
		}

		return nil
	})
}

func shouldSkipFile(info fs.FileInfo) bool {
	// skip if not a file
	if info.IsDir() {
		return true
	}

	// skip if not a .yml or .yaml file
	if !strings.HasSuffix(info.Name(), ".yml") && !strings.HasSuffix(info.Name(), ".yaml") {
		return true
	}

	return false
}
