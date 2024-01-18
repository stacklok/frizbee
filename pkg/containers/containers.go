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

// Package containers provides functions to replace tags for checksums
package containers

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stacklok/frizbee/pkg/constants"
)

// GetDigest returns the digest of a container image reference.
func GetDigest(ctx context.Context, refstr string, opts ...remote.Option) (string, error) {
	ref, err := name.ParseReference(refstr)
	if err != nil {
		return "", fmt.Errorf("failed to parse reference: %w", err)
	}

	return GetDigestFromRef(ctx, ref, opts...)
}

// GetDigestFromRef returns the digest of a container image reference
// from a name.Reference.
func GetDigestFromRef(ctx context.Context, ref name.Reference, opts ...remote.Option) (string, error) {
	opts = append(opts, remote.WithContext(ctx), remote.WithUserAgent(constants.UserAgent))
	desc, err := remote.Get(ref, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to get remote reference: %w", err)
	}

	// check if descriptor points to an index and return the digest if it does
	// if desc.MediaType.IsIndex() {
	// 	return desc.Digest.String(), nil
	// }

	img, err := desc.Image()
	if err != nil {
		return "", err
	}
	digest, err := img.Digest()
	if err != nil {
		return "", err
	}
	return digest.String(), nil
}
