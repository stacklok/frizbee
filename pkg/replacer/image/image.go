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

package image

import (
	"context"
	"fmt"
	"github.com/stacklok/frizbee/internal/store"
	"github.com/stacklok/frizbee/pkg/config"
	ferrors "github.com/stacklok/frizbee/pkg/errors"
	"github.com/stacklok/frizbee/pkg/interfaces"
	"strings"
)

const (
	// ContainerImageRegex is regular expression pattern to match container image usage in YAML
	ContainerImageRegex = `image\s*:\s*["']?([^\s"']+/[^\s"']+|[^\s"']+)(:[^\s"']+)?(@[^\s"']+)?["']?|FROM\s+([^\s]+(/[^\s]+)?(:[^\s]+)?(@[^\s]+)?)` // `image\s*:\s*["']?([^\s"']+/[^\s"']+(:[^\s"']+)?(@[^\s"']+)?)["']?|FROM\s+([^\s]+(/[^\s]+)?(:[^\s]+)?(@[^\s]+)?)` // `\b(image|FROM)\s*:?(\s*([^\s]+))?`
	prefixFROM          = "FROM "
	prefixImage         = "image: "
	ReferenceType       = "container"
)

type Parser struct {
	regex string
}

func New(regex string) *Parser {
	if regex == "" {
		regex = ContainerImageRegex
	}
	return &Parser{
		regex: regex,
	}
}

func (p *Parser) GetRegex() string {
	return p.regex
}

func (p *Parser) Replace(ctx context.Context, matchedLine string, _ interfaces.REST, cfg config.Config, cache store.RefCacher) (*interfaces.EntityRef, error) {
	// Trim the prefix
	hasFROMPrefix := false

	// Check if the image reference has the FROM prefix, i.e. Dockerfile
	if strings.HasPrefix(matchedLine, prefixFROM) {
		matchedLine = strings.TrimPrefix(matchedLine, prefixFROM)
		// Check if the image reference should be excluded, i.e. scratch
		if shouldExclude(matchedLine) {
			return nil, fmt.Errorf("image reference %s should be excluded - %w", matchedLine, ferrors.ErrReferenceSkipped)
		}
		hasFROMPrefix = true
	} else if strings.HasPrefix(matchedLine, prefixImage) {
		// Check if the image reference has the image prefix, i.e. Kubernetes or Docker Compose YAML
		matchedLine = strings.TrimPrefix(matchedLine, prefixImage)
	}

	// Get the digest of the image reference
	imageRefWithDigest, err := GetImageDigestFromRef(ctx, matchedLine, cfg.Platform, cache)
	if err != nil {
		return nil, err
	}

	// Add the prefix back
	if hasFROMPrefix {
		imageRefWithDigest.Prefix = fmt.Sprintf("%s%s", prefixFROM, imageRefWithDigest.Prefix)
	} else {
		imageRefWithDigest.Prefix = fmt.Sprintf("%s%s", prefixImage, imageRefWithDigest.Prefix)
	}

	// Return the reference
	return imageRefWithDigest, nil
}

func (p *Parser) ConvertToEntityRef(reference string) (*interfaces.EntityRef, error) {
	reference = strings.TrimPrefix(reference, prefixImage)
	reference = strings.TrimPrefix(reference, prefixFROM)
	var sep string
	var frags []string
	if strings.Contains(reference, "@") {
		sep = "@"
	} else if strings.Contains(reference, ":") {
		sep = ":"
	}

	if sep != "" {
		frags = strings.Split(reference, sep)
		if len(frags) != 2 {
			return nil, fmt.Errorf("invalid container reference: %s", reference)
		}
	} else {
		frags = []string{reference, "latest"}
	}

	return &interfaces.EntityRef{
		Name: frags[0],
		Ref:  frags[1],
		Type: ReferenceType,
	}, nil
}
