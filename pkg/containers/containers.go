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

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	gocrv1 "github.com/google/go-containerregistry/pkg/v1"
)

// nolint: gochecknoglobals
var platform *gocrv1.Platform

// SetPlatform sets the platform to be used for getting digests.
func SetPlatform(pf *gocrv1.Platform) {
	platform = pf
}

// GetDigest returns the digest of a container image reference.
func GetDigest(ctx context.Context, refstr string) (string, error) {
	ref, err := name.ParseReference(refstr)
	if err != nil {
		return "", fmt.Errorf("failed to parse reference: %w", err)
	}

	return GetDigestFromRef(ctx, ref)
}

// GetDigestFromRef returns the digest of a container image reference
// from a name.Reference.
func GetDigestFromRef(ctx context.Context, ref name.Reference) (string, error) {
	digest, err := crane.Digest(ref.String(),
		crane.WithContext(ctx),
		crane.WithPlatform(platform),
	)
	if err != nil {
		return "", fmt.Errorf("failed to get remote reference: %w", err)
	}

	return digest, nil
}
