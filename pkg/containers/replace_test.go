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

package containers

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceImageReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name           string
		input          string
		expectedOutput string
		modified       bool
	}{
		{
			name: "Replace image reference",
			input: `
version: v1
services:
  - name: web
    image: nginx:1.25.3
  - name: localstack
    image: localstack/localstack:3.0.2
`,
			expectedOutput: `
version: v1
services:
  - name: web
    image: index.docker.io/library/nginx@sha256:2bdc49f2f8ae8d8dc50ed00f2ee56d00385c6f8bc8a8b320d0a294d9e3b49026  # 1.25.3
  - name: localstack
    image: index.docker.io/localstack/localstack@sha256:e606c4421419030b12d63a59f1211f57f5b0fbf7e9ce769e6250ee62ff4f9293  # 3.0.2
`,
			modified: true,
		},
		// Add more test cases as needed
	}

	// Define a regular expression to match YAML tags containing "image"
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var output strings.Builder
			m, err := ReplaceImageReferenceFromYAML(ctx, strings.NewReader(tc.input), &output)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedOutput, output.String())
			assert.Equal(t, tc.modified, m, "modified")
		})
	}
}
