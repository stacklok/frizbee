package image

import (
	"context"
	"fmt"
	"github.com/stacklok/frizbee/internal/store"
	"github.com/stacklok/frizbee/pkg/config"
	"github.com/stacklok/frizbee/pkg/interfaces"
	"strings"
)

const (
	// ContainerImageRegex is regular expression pattern to match container image usage in YAML
	ContainerImageRegex = `image\s*:\s*["']?([^\s"']+/[^\s"']+(:[^\s"']+)?(@[^\s"']+)?)["']?|FROM\s+([^\s]+(/[^\s]+)?(:[^\s]+)?(@[^\s]+)?)` // `\b(image|FROM)\s*:?(\s*([^\s]+))?`
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

func (p *Parser) Replace(ctx context.Context, matchedLine string, _ interfaces.REST, cfg config.Config, cache store.RefCacher, keepPrefix bool) (string, error) {
	// Trim the prefix
	hasFROMPrefix := false
	imageRef := matchedLine

	// Check if the image reference has the FROM prefix, i.e. Dockerfile
	if strings.HasPrefix(imageRef, prefixFROM) {
		imageRef = strings.TrimPrefix(imageRef, prefixFROM)
		// Check if the image reference should be excluded, i.e. scratch
		if shouldExclude(imageRef) {
			return matchedLine, nil
		}
		hasFROMPrefix = true
	} else if strings.HasPrefix(imageRef, prefixImage) {
		// Check if the image reference has the image prefix, i.e. Kubernetes or Docker Compose YAML
		imageRef = strings.TrimPrefix(imageRef, prefixImage)
	}

	// Get the digest of the image reference
	imageRefWithDigest, err := GetImageDigestFromRef(ctx, imageRef, cfg.Platform, cache, hasFROMPrefix)
	if err != nil {
		return "", err
	}

	// Add the prefix back, if needed
	if keepPrefix {
		if hasFROMPrefix {
			imageRefWithDigest = prefixFROM + imageRefWithDigest
		} else {
			imageRefWithDigest = prefixImage + imageRefWithDigest
		}
		// Return the modified line with the prefix
		return imageRefWithDigest, nil

	}
	// Return the modified line without the prefix
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
