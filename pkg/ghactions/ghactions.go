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

// Package ghactions provides functions to replace tags for checksums
// in GitHub Actions workflows.
package ghactions

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v56/github"
	"gopkg.in/yaml.v3"
)

// IsLocal returns true if the input is a local path.
func IsLocal(input string) bool {
	return strings.HasPrefix(input, "./") || strings.HasPrefix(input, "../")
}

// ParseActionReference parses an action reference into action and reference.
func ParseActionReference(input string) (action string, reference string, err error) {
	frags := strings.Split(input, "@")
	if len(frags) != 2 {
		return "", "", fmt.Errorf("invalid action reference: %s", input)
	}

	return frags[0], frags[1], nil
}

// GetChecksum returns the checksum for a given action and tag.
func GetChecksum(ctx context.Context, ghcli *github.Client, action, ref string) (string, error) {
	owner, repo, err := parseActionFragments(action)
	if err != nil {
		return "", err
	}

	res, err := getCheckSumForTag(ctx, ghcli, owner, repo, ref)
	if err != nil {
		return "", fmt.Errorf("failed to get checksum for tag: %w", err)
	} else if res != "" {
		return res, nil
	}

	// check branch
	res, err = getCheckSumForBranch(ctx, ghcli, owner, repo, ref)
	if err != nil {
		return "", fmt.Errorf("failed to get checksum for branch: %w", err)
	} else if res != "" {
		return res, nil
	}

	// Check if we're using a checksum
	if len(ref) != 40 {
		return "", fmt.Errorf("given reference is not a tag nor branch")
	}

	return ref, nil
}

// ModifyReferencesInYAML takes the given YAML structure and replaces
// all references to tags with the checksum of the tag.
// Note that the given YAML structure is modified in-place.
// The function returns true if any references were modified.
func ModifyReferencesInYAML(ctx context.Context, ghcli *github.Client, node *yaml.Node) (bool, error) {
	// `uses` will be immediately before the action
	// name in the YAML `Content` array. We use a toggle
	// to track if we've found `uses` and then look for
	// the next node.
	foundUses := false
	modified := false

	for _, v := range node.Content {
		if v.Value == "uses" {
			foundUses = true
			continue
		}

		if foundUses {
			foundUses = false

			// If the value is a local path, skip it
			if IsLocal(v.Value) {
				continue
			}

			act, ref, err := ParseActionReference(v.Value)
			if err != nil {
				return modified, fmt.Errorf("failed to parse action reference '%s': %w", v.Value, err)
			}

			sum, err := GetChecksum(ctx, ghcli, act, ref)
			if err != nil {
				return modified, fmt.Errorf("failed to get checksum for action '%s': %w", v.Value, err)
			}

			if ref != sum {
				v.SetString(fmt.Sprintf("%s@%s", act, sum))
				v.LineComment = ref
				modified = true
			}
			continue
		}

		// Otherwise recursively look more
		m, err := ModifyReferencesInYAML(ctx, ghcli, v)
		if err != nil {
			return m, err
		}
		modified = modified || m
	}
	return modified, nil
}
