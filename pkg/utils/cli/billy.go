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

// Package cli provides utilities for frizbee's CLI.
package cli

// ProcessDirNameForBillyFS processes the given directory name for use with
// go-billy filesystems.
func ProcessDirNameForBillyFS(dir string) string {
	// remove trailing / from dir. This doesn't play well with
	// the go-billy filesystem and walker we use.
	if dir[len(dir)-1] == '/' {
		return dir[:len(dir)-1]
	}

	return dir
}
