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

package utils

import (
	"fmt"
	"strings"

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
	defer enc.Close()

	return &buf, nil
}
