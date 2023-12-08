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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-github/v56/github"

	"github.com/stacklok/frizbee/pkg/interfaces"
)

func parseActionFragments(action string) (owner string, repo string, err error) {
	frags := strings.Split(action, "/")

	// if we have more than 2 fragments, we're probably dealing with
	// sub-actions, so we take the first two fragments as the owner and repo
	if len(frags) < 2 {
		return "", "", fmt.Errorf("%w: '%s' reference is incorrect", ErrInvalidAction, action)
	}

	return frags[0], frags[1], nil
}

func getCheckSumForTag(ctx context.Context, restIf interfaces.REST, owner, repo, tag string) (string, error) {
	path, err := url.JoinPath("repos", owner, repo, "git", "refs", "tags", tag)
	if err != nil {
		return "", fmt.Errorf("failed to join path: %w", err)
	}

	return doGetReference(ctx, restIf, path)
}

func getCheckSumForBranch(ctx context.Context, restIf interfaces.REST, owner, repo, branch string) (string, error) {
	path, err := url.JoinPath("repos", owner, repo, "git", "refs", "heads", branch)
	if err != nil {
		return "", fmt.Errorf("failed to join path: %w", err)
	}

	return doGetReference(ctx, restIf, path)
}

func doGetReference(ctx context.Context, restIf interfaces.REST, path string) (string, error) {
	req, err := restIf.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create REST request: %w", err)
	}

	resp, err := restIf.Do(ctx, req)

	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}

	if err != nil && resp.StatusCode != http.StatusNotFound {
		return "", fmt.Errorf("failed to do API request: %w", err)
	} else if resp.StatusCode == http.StatusNotFound {
		// No error, but no tag found
		return "", nil
	}

	var t github.Reference
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil && strings.Contains(err.Error(), "cannot unmarshal array into Go value of type") {
		// This is a branch, not a tag
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("canont decode response: %w", err)
	}

	return t.GetObject().GetSHA(), nil
}
