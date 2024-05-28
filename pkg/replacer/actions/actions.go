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

// Package actions provides utilities to work with GitHub Actions.
package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/stacklok/frizbee/pkg/interfaces"
	"github.com/stacklok/frizbee/pkg/replacer/image"
	"github.com/stacklok/frizbee/pkg/utils/config"
	"github.com/stacklok/frizbee/pkg/utils/store"
)

const (
	// GitHubActionsRegex is regular expression pattern to match GitHub Actions usage
	GitHubActionsRegex = `uses:\s*[^\s]+/[^\s]+@[^\s]+|uses:\s*docker://[^\s]+:[^\s]+`
	prefixUses         = "uses: "
	prefixDocker       = "docker://"
	// ReferenceType is the type of the reference
	ReferenceType = "action"
)

// Parser is a struct to replace action references with digests
type Parser struct {
	regex string
	cache store.RefCacher
}

// New creates a new Parser
func New() *Parser {
	return &Parser{
		regex: GitHubActionsRegex,
		cache: store.NewRefCacher(),
	}
}

// SetCache returns the regular expression pattern to match GitHub Actions usage
func (p *Parser) SetCache(cache store.RefCacher) {
	p.cache = cache
}

// SetRegex returns the regular expression pattern to match GitHub Actions usage
func (p *Parser) SetRegex(regex string) {
	p.regex = regex
}

// GetRegex returns the regular expression pattern to match GitHub Actions usage
func (p *Parser) GetRegex() string {
	return p.regex
}

// Replace replaces the action reference with the digest
func (p *Parser) Replace(
	ctx context.Context,
	matchedLine string,
	restIf interfaces.REST,
	cfg config.Config,
) (*interfaces.EntityRef, error) {
	var err error
	var actionRef *interfaces.EntityRef
	hasUsesPrefix := false

	// Trim the uses prefix
	if strings.HasPrefix(matchedLine, prefixUses) {
		matchedLine = strings.TrimPrefix(matchedLine, prefixUses)
		hasUsesPrefix = true
	}
	// Determine if the action reference has a docker prefix
	if strings.HasPrefix(matchedLine, prefixDocker) {
		actionRef, err = p.replaceDocker(ctx, matchedLine, restIf, cfg)
	} else {
		actionRef, err = p.replaceAction(ctx, matchedLine, restIf, cfg)
	}
	if err != nil {
		return nil, err
	}

	// Add back the uses prefix
	if hasUsesPrefix {
		actionRef.Prefix = fmt.Sprintf("%s%s", prefixUses, actionRef.Prefix)
	}

	// Return the new action reference
	return actionRef, nil
}

func (p *Parser) replaceAction(
	ctx context.Context,
	matchedLine string,
	restIf interfaces.REST,
	cfg config.Config,
) (*interfaces.EntityRef, error) {

	// If the value is a local path or should be excluded, skip it
	if isLocal(matchedLine) || shouldExclude(&cfg.GHActions, matchedLine) {
		return nil, fmt.Errorf("%w: %s", interfaces.ErrReferenceSkipped, matchedLine)
	}

	// Parse the action reference
	act, ref, err := ParseActionReference(matchedLine)
	if err != nil {
		return nil, fmt.Errorf("failed to parse action reference '%s': %w", matchedLine, err)
	}

	// Check if the parsed reference should be excluded
	if shouldExclude(&cfg.GHActions, act) {
		return nil, fmt.Errorf("%w: %s", interfaces.ErrReferenceSkipped, matchedLine)
	}
	var sum string

	// Check if we have a cache
	if p.cache != nil {
		// Check if we have a cached value
		if val, ok := p.cache.Load(matchedLine); ok {
			sum = val
		} else {
			// Get the checksum for the action reference
			sum, err = GetChecksum(ctx, restIf, act, ref)
			if err != nil {
				return nil, fmt.Errorf("failed to get checksum for action '%s': %w", matchedLine, err)
			}
			// Store the checksum in the cache
			p.cache.Store(matchedLine, sum)
		}
	} else {
		// Get the checksum for the action reference
		sum, err = GetChecksum(ctx, restIf, act, ref)
		if err != nil {
			return nil, fmt.Errorf("failed to get checksum for action '%s': %w", matchedLine, err)
		}
	}

	// Compare the digest with the reference and return the original reference if they already match
	if ref == sum {
		return nil, fmt.Errorf("image already referenced by digest: %s %w", matchedLine, interfaces.ErrReferenceSkipped)
	}

	return &interfaces.EntityRef{
		Name: act,
		Ref:  sum,
		Type: ReferenceType,
		Tag:  ref,
	}, nil
}

func (p *Parser) replaceDocker(
	ctx context.Context,
	matchedLine string,
	_ interfaces.REST,
	cfg config.Config,
) (*interfaces.EntityRef, error) {
	// Trim the docker prefix
	trimmedRef := strings.TrimPrefix(matchedLine, prefixDocker)

	// If the value is a local path or should be excluded, skip it
	if isLocal(trimmedRef) || shouldExclude(&cfg.GHActions, trimmedRef) {
		return nil, fmt.Errorf("%w: %s", interfaces.ErrReferenceSkipped, matchedLine)
	}

	// Get the digest of the docker:// image reference
	actionRef, err := image.GetImageDigestFromRef(ctx, trimmedRef, cfg.Platform, p.cache)
	if err != nil {
		return nil, err
	}

	// Check if the parsed reference should be excluded
	if shouldExclude(&cfg.GHActions, actionRef.Name) {
		return nil, fmt.Errorf("%w: %s", interfaces.ErrReferenceSkipped, matchedLine)
	}

	// Add back the docker prefix
	if strings.HasPrefix(matchedLine, prefixDocker) {
		actionRef.Prefix = fmt.Sprintf("%s%s", prefixDocker, actionRef.Prefix)
	}

	return actionRef, nil
}

// ConvertToEntityRef converts an action reference to an EntityRef
func (_ *Parser) ConvertToEntityRef(reference string) (*interfaces.EntityRef, error) {
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
