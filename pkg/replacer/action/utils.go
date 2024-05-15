package action

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-github/v61/github"
	"github.com/stacklok/frizbee/pkg/config"
	"github.com/stacklok/frizbee/pkg/interfaces"
	"net/http"
	"net/url"
	"strings"
)

var (
	// ErrInvalidAction is returned when parsing the action fails.
	ErrInvalidAction = errors.New("invalid action")

	// ErrInvalidActionReference is returned when parsing the action reference fails.
	ErrInvalidActionReference = errors.New("action reference is not a tag nor branch")
)

// isLocal returns true if the input is a local path.
func isLocal(input string) bool {
	return strings.HasPrefix(input, "./") || strings.HasPrefix(input, "../")
}

func shouldExclude(cfg *config.GHActions, input string) bool {
	for _, e := range cfg.Exclude {
		if e == input {
			return true
		}
	}
	return false
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
func GetChecksum(ctx context.Context, restIf interfaces.REST, action, ref string) (string, error) {
	owner, repo, err := parseActionFragments(action)
	if err != nil {
		return "", err
	}

	// Check if we're using a checksum
	if isChecksum(ref) {
		return ref, nil
	}

	res, err := getCheckSumForTag(ctx, restIf, owner, repo, ref)
	if err != nil {
		return "", fmt.Errorf("failed to get checksum for tag: %w", err)
	} else if res != "" {
		return res, nil
	}

	// check branch
	res, err = getCheckSumForBranch(ctx, restIf, owner, repo, ref)
	if err != nil {
		return "", fmt.Errorf("failed to get checksum for branch: %w", err)
	} else if res != "" {
		return res, nil
	}

	return "", ErrInvalidActionReference
}

func parseActionFragments(action string) (owner string, repo string, err error) {
	frags := strings.Split(action, "/")

	// if we have more than 2 fragments, we're probably dealing with
	// sub-actions, so we take the first two fragments as the owner and repo
	if len(frags) < 2 {
		return "", "", fmt.Errorf("%w: '%s' reference is incorrect", ErrInvalidAction, action)
	}

	return frags[0], frags[1], nil
}

// isChecksum returns true if the input is a checksum.
func isChecksum(ref string) bool {
	return len(ref) == 40
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
