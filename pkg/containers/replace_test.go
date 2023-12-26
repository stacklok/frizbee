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
  - name: kube-apiserver
    image: registry.k8s.io/kube-apiserver:v1.20.0
  - name: kube-controller-manager
    image: registry.k8s.io/kube-controller-manager:v1.15.0
`,
			expectedOutput: `
version: v1
services:
  - name: kube-apiserver
    image: registry.k8s.io/kube-apiserver@sha256:8b8125d7a6e4225b08f04f65ca947b27d0cc86380bf09fab890cc80408230114  # v1.20.0
  - name: kube-controller-manager
    image: registry.k8s.io/kube-controller-manager@sha256:835f32a5cdb30e86f35675dd91f9c7df01d48359ab8b51c1df866a2c7ea2e870  # v1.15.0
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
