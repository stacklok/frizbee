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
    image: nginx:latest
  - name: localstack
    image: localstack/localstack
`,
			expectedOutput: `
version: v1
services:
  - name: web
    image: index.docker.io/library/nginx@sha256:10d1f5b58f74683ad34eb29287e07dab1e90f10af243f151bb50aa5dbb4d62ee  # latest
  - name: localstack
    image: index.docker.io/localstack/localstack@sha256:9b89e7d3bd1b0869f58d9aff0bfad30b4e1c2491ece7a00fb0a7515530d69cf2  # latest
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
