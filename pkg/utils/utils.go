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

// Package utils provides utilities for frizbee
package utils

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/go-git/go-billy/v5"
	billyutil "github.com/go-git/go-billy/v5/util"
	"gopkg.in/yaml.v3"
)

// YAMLToBuffer converts a YAML node to a string buffer
func YAMLToBuffer(wflow *yaml.Node) (fmt.Stringer, error) {
	buf := strings.Builder{}
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(wflow); err != nil {
		return nil, fmt.Errorf("failed to encode YAML: %w", err)
	}

	// nolint:errcheck // ignore error
	defer enc.Close()

	return &buf, nil
}

// TraverseFunc is a function that gets called with each file in a directory.
type TraverseFunc func(path string, info fs.FileInfo) error

// Traverse traverses the given directory and calls the given function with each file.
func Traverse(bfs billy.Filesystem, base string, fun TraverseFunc) error {
	return billyutil.Walk(bfs, base, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		return fun(path, info)
	})
}

// IsYAMLFile returns true if the given file is a YAML file.
func IsYAMLFile(info fs.FileInfo) bool {
	// skip if not a file
	if info.IsDir() {
		return false
	}

	// skip if not a .yml or .yaml file
	if strings.HasSuffix(info.Name(), ".yml") || strings.HasSuffix(info.Name(), ".yaml") {
		return true
	}

	return false
}
