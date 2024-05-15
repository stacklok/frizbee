package image

import (
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stacklok/frizbee/internal/cli"
	"github.com/stacklok/frizbee/internal/store"
	"strings"
)

// GetImageDigestFromRef returns the digest of a container image reference
// from a name.Reference.
func GetImageDigestFromRef(ctx context.Context, imageRef, platform string, cache store.RefCacher, isDockerfileRef bool) (string, error) {
	// Parse the image reference
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return "", err
	}
	opts := []remote.Option{
		remote.WithContext(ctx),
		remote.WithUserAgent(cli.UserAgent),
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	}

	// Set the platform if provided
	if platform != "" {
		platformSplit := strings.Split(platform, "/")
		if len(platformSplit) != 2 {
			return "", fmt.Errorf("platform must be in the format os/arch")
		}
		opts = append(opts, remote.WithPlatform(v1.Platform{
			OS:           platformSplit[0],
			Architecture: platformSplit[1],
		}))
	}

	// Get the digest of the image reference
	var digest string

	if cache != nil {
		if d, ok := cache.Load(imageRef); ok {
			digest = d
		}
		desc, err := remote.Get(ref, opts...)
		if err != nil {
			return "", err
		}
		digest = desc.Digest.String()
		cache.Store(imageRef, digest)
	} else {
		desc, err := remote.Get(ref, opts...)
		if err != nil {
			return "", err
		}
		digest = desc.Digest.String()
	}

	// Compare the digest with the reference and return the original reference if they already match
	if digest == ref.Identifier() {
		return imageRef, nil
	}

	// Return the image reference with the digest differently if it is a Dockerfile reference
	if isDockerfileRef {
		return fmt.Sprintf("%s:%s@%s", ref.Context().Name(), ref.Identifier(), digest), nil
	}
	return fmt.Sprintf("%s@%s # %s", ref.Context().Name(), digest, ref.Identifier()), nil
}

func shouldExclude(ref string) bool {
	return ref == "scratch"
}
