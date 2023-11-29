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
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"

	"github.com/google/go-containerregistry/pkg/name"
)

// ReplaceImageReferenceFromYAML replaces the image reference in the input text with the digest
func ReplaceImageReferenceFromYAML(ctx context.Context, input io.Reader, output io.Writer) (bool, error) {
	return ReplaceReferenceFromYAML(ctx, "image", input, output)
}

// ReplaceReferenceFromYAML replaces the image reference in the input text with the digest
func ReplaceReferenceFromYAML(ctx context.Context, keyRegex string, input io.Reader, output io.Writer) (bool, error) {
	scanner := bufio.NewScanner(input)
	re, err := regexp.Compile(fmt.Sprintf(`(\s*%s):\s*([^\s]+)`, keyRegex))
	if err != nil {
		return false, fmt.Errorf("failed to compile regex: %w", err)
	}

	modified := false

	for scanner.Scan() {
		line := scanner.Text()
		updatedLine := re.ReplaceAllStringFunc(line, func(match string) string {
			submatches := re.FindStringSubmatch(match)
			if len(submatches) != 3 {
				return match
			}

			imageReferenceWithTag := submatches[2]
			ref, err := name.ParseReference(imageReferenceWithTag)
			if err != nil {
				return match
			}

			digest, err := GetDigestFromRef(ctx, ref)
			if err != nil {
				return match
			}

			imgWithoutTag := ref.Context().Name()
			outstr := imgWithoutTag + "@" + digest

			if imageReferenceWithTag != outstr {
				modified = true
			}

			replacement := fmt.Sprintf("${1}: %s  # %s", outstr, ref.Identifier())
			return re.ReplaceAllString(match, replacement)
		})

		if _, err := io.WriteString(output, updatedLine+"\n"); err != nil {
			return false, err
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return modified, nil
}
