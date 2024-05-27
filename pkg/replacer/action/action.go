//
// Copyright 2024 Stacklok, Inc.
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

package action

import (
	"context"
	"fmt"
	"github.com/stacklok/frizbee/internal/store"
	"github.com/stacklok/frizbee/pkg/config"
	"github.com/stacklok/frizbee/pkg/interfaces"
	"github.com/stacklok/frizbee/pkg/replacer/image"
	"strings"
)

const (
	// GitHubActionsRegex is regular expression pattern to match GitHub Actions usage
	GitHubActionsRegex = `uses:\s*[^\s]+/[^\s]+@[^\s]+|uses:\s*docker://[^\s]+:[^\s]+`
	prefixUses         = "uses: "
	prefixDocker       = "docker://"
	ReferenceType      = "action"
)

type Parser struct {
	regex string
}

func New(regex string) *Parser {
	if regex == "" {
		regex = GitHubActionsRegex
	}
	return &Parser{
		regex: regex,
	}
}

func (p *Parser) GetRegex() string {
	return p.regex
}

func (p *Parser) Replace(ctx context.Context, matchedLine string, restIf interfaces.REST, cfg config.Config, cache store.RefCacher, keepPrefix bool) (string, error) {
	var err error

	// Trim the uses prefix
	actionRef := strings.TrimPrefix(matchedLine, prefixUses)

	// Determine if the action reference has a docker prefix
	if strings.Contains(actionRef, prefixDocker) {
		actionRef, err = p.replaceDocker(ctx, actionRef, restIf, cfg, cache, keepPrefix)
	} else {
		actionRef, err = p.replaceAction(ctx, actionRef, restIf, cfg, cache, keepPrefix)
	}
	if err != nil {
		return "", err
	}

	// Add back the uses prefix, if needed
	if keepPrefix {
		actionRef = fmt.Sprintf("%s%s", prefixUses, actionRef)
	}

	// Return the new action reference
	return actionRef, nil
}

func (p *Parser) replaceAction(ctx context.Context, matchedLine string, restIf interfaces.REST, cfg config.Config, cache store.RefCacher, keepPrefix bool) (string, error) {
	actionRef := matchedLine

	// If the value is a local path or should be excluded, skip it
	if isLocal(actionRef) || shouldExclude(&cfg.GHActions, actionRef) {
		return matchedLine, nil
	}

	// Parse the action reference
	act, ref, err := ParseActionReference(actionRef)
	if err != nil {
		return matchedLine, nil
	}

	// Check if the parsed reference should be excluded
	if shouldExclude(&cfg.GHActions, act) {
		return matchedLine, nil
	}
	var sum string

	// Check if we have a cache
	if cache != nil {
		// Check if we have a cached value
		if val, ok := cache.Load(actionRef); ok {
			sum = val
		} else {
			// Get the checksum for the action reference
			sum, err = GetChecksum(ctx, restIf, act, ref)
			if err != nil {
				return matchedLine, nil
			}
			// Store the checksum in the cache
			cache.Store(actionRef, sum)
		}
	} else {
		// Get the checksum for the action reference
		sum, err = GetChecksum(ctx, restIf, act, ref)
		if err != nil {
			return matchedLine, nil
		}
	}
	// If the checksum is different from the reference, update the reference
	// Otherwise, return the original line
	if ref == sum {
		return matchedLine, nil
	}

	return fmt.Sprintf("%s@%s # %s", act, sum, ref), nil
}

func (p *Parser) replaceDocker(ctx context.Context, matchedLine string, _ interfaces.REST, cfg config.Config, cache store.RefCacher, keepPrefix bool) (string, error) {
	var err error
	// Trim the docker prefix
	actionRef := strings.TrimPrefix(matchedLine, prefixDocker)

	// If the value is a local path or should be excluded, skip it
	if isLocal(actionRef) || shouldExclude(&cfg.GHActions, actionRef) {
		return matchedLine, nil
	}

	// Get the digest of the docker:// image reference
	actionRef, err = image.GetImageDigestFromRef(ctx, actionRef, cfg.Platform, cache, false)
	if err != nil {
		return "", err
	}

	// Check if the parsed reference should be excluded
	if shouldExclude(&cfg.GHActions, actionRef) {
		return matchedLine, nil
	}

	// Add back the docker prefix, if needed
	if keepPrefix {
		actionRef = fmt.Sprintf("%s%s", prefixDocker, actionRef)
	}
	return actionRef, nil
}

func (p *Parser) ConvertToEntityRef(reference string) (*interfaces.EntityRef, error) {
	reference = strings.TrimPrefix(reference, prefixUses)
	refType := ReferenceType
	separator := "@"
	// Update the separator in case this is a docker reference with a digest
	if strings.Contains(reference, prefixDocker) {
		reference = strings.TrimPrefix(reference, prefixDocker)
		if !strings.Contains(reference, separator) && strings.Contains(reference, ":") {
			separator = ":"
		}
		refType = image.ReferenceType
	}
	frags := strings.Split(reference, separator)
	if len(frags) != 2 {
		return nil, fmt.Errorf("invalid action reference: %s", reference)
	}

	return &interfaces.EntityRef{
		Name: frags[0],
		Ref:  frags[1],
		Type: refType,
	}, nil
}
