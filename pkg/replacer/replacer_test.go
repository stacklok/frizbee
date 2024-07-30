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

package replacer

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/require"

	"github.com/stacklok/frizbee/internal/cli"
	"github.com/stacklok/frizbee/pkg/interfaces"
	"github.com/stacklok/frizbee/pkg/replacer/actions"
	"github.com/stacklok/frizbee/pkg/replacer/image"
	"github.com/stacklok/frizbee/pkg/utils/config"
	"github.com/stacklok/frizbee/pkg/utils/ghrest"
)

func TestReplacer_ParseContainerImageString(t *testing.T) {
	t.Parallel()

	type args struct {
		refstr string
	}
	tests := []struct {
		name    string
		args    args
		want    *interfaces.EntityRef
		wantErr bool
	}{
		{
			name: "dockerfile - tag",
			args: args{
				refstr: "FROM golang:1.22.2",
			},
			want: &interfaces.EntityRef{
				Name:   "index.docker.io/library/golang",
				Ref:    "sha256:d5302d40dc5fbbf38ec472d1848a9d2391a13f93293a6a5b0b87c99dc0eaa6ae",
				Type:   image.ReferenceType,
				Tag:    "1.22.2",
				Prefix: "FROM ",
			},
			wantErr: false,
		},
		{
			name: "dockerfile - tag, stage and platform",
			args: args{
				refstr: "FROM --platform=linux/s390x golang:1.22.2 AS build",
			},
			want: &interfaces.EntityRef{
				Name:   "index.docker.io/library/golang",
				Ref:    "sha256:d5302d40dc5fbbf38ec472d1848a9d2391a13f93293a6a5b0b87c99dc0eaa6ae",
				Type:   image.ReferenceType,
				Tag:    "1.22.2",
				Prefix: "FROM --platform=linux/s390x ",
			},
			wantErr: false,
		},
		{
			name: "dockerfile - no tag",
			args: args{
				refstr: "FROM golang",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "dockerfile - latest",
			args: args{
				refstr: "FROM golang:latest",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "dockerfile - already by digest",
			args: args{
				refstr: "FROM golang:1.22.2@sha256:aca60c1f21de99aa3a34e653f0cdc8c8ea8fe6480359229809d5bcb974f599ec",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "dockerfile - scratch",
			args: args{
				refstr: "FROM scratch",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid 1",
			args: args{
				refstr: "ghcr.io/stacklok/minder/helm/minder:0.20231123.829_ref.26ca90b",
			},
			want: &interfaces.EntityRef{
				Name:   "ghcr.io/stacklok/minder/helm/minder",
				Ref:    "sha256:a29f8a8d28f0af7f70a4b3dd3e33c8c8cc5cf9e88e802e2700cf272a0b6140ec",
				Type:   image.ReferenceType,
				Tag:    "0.20231123.829_ref.26ca90b",
				Prefix: "",
			},
			wantErr: false,
		},
		{
			name: "valid 2",
			args: args{
				refstr: "devopsfaith/krakend:2.5.0",
			},
			want: &interfaces.EntityRef{
				Name:   "index.docker.io/devopsfaith/krakend",
				Ref:    "sha256:6a3c8e5e1a4948042bfb364ed6471e16b4a26d0afb6c3c01ebcb88b3fa551036",
				Type:   image.ReferenceType,
				Tag:    "2.5.0",
				Prefix: "",
			},
			wantErr: false,
		},
		{
			name: "image with no tag is skipped",
			args: args{
				refstr: "image: nginx",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "image with latest tag is skipped",
			args: args{
				refstr: "image: nginx:latest",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid ref string",
			args: args{
				refstr: "ghcr.io/stacklok/minder/helm/minder!",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "nonexistent container in nonexistent registry",
			args: args{
				refstr: "beeeeer.io/ipa/toppling-goliath/king-sue:1.0.0",
			},
			want:    nil,
			wantErr: true,
		},
		// TODO: Create a dedicated container image for this test and push it so that latest doesnt change
		//{
		//	name: "container reference with no tag or digest",
		//	args: args{
		//		refstr: "nginx",
		//	},
		//	want: &interfaces.EntityRef{
		//		Name:   "index.docker.io/library/nginx",
		//		Ref:    "sha256:faef0b115e699b1e70b1f9a939ea2bc62c26485f6b72e91c8a7b236f1f8589c1",
		//		Type:   image.ReferenceType,
		//		Tag:    "latest",
		//		Prefix: "",
		//	},
		//	wantErr: false,
		//},
		{
			name: "invalid reference with special characters",
			args: args{
				refstr: "nginx@#$$%%^&*",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			config := &config.Config{
				Images: config.Images{
					ImageFilter: config.ImageFilter{
						ExcludeTags: []string{"latest"},
					},
				},
			}
			r := NewContainerImagesReplacer(config)
			got, err := r.ParseString(ctx, tt.args.refstr)
			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestReplacer_ParseGitHubActionString(t *testing.T) {
	t.Parallel()

	type args struct {
		action string
	}
	tests := []struct {
		name    string
		args    args
		want    *interfaces.EntityRef
		wantErr bool
	}{
		{
			name: "action using a container via docker://avtodev/markdown-lint:v1",
			args: args{
				action: "uses: docker://avtodev/markdown-lint:v1",
			},
			want: &interfaces.EntityRef{
				Name:   "index.docker.io/avtodev/markdown-lint",
				Ref:    "sha256:6aeedc2f49138ce7a1cd0adffc1b1c0321b841dc2102408967d9301c031949ee",
				Type:   image.ReferenceType,
				Tag:    "v1",
				Prefix: "uses: docker://",
			},
			wantErr: false,
		},
		{
			name: "actions/checkout with v4.1.1",
			args: args{
				action: "actions/checkout@v4.1.1",
			},
			want: &interfaces.EntityRef{
				Name:   "actions/checkout",
				Ref:    "b4ffde65f46336ab88eb53be808477a3936bae11",
				Type:   actions.ReferenceType,
				Tag:    "v4.1.1",
				Prefix: "",
			},
			wantErr: false,
		},
		{
			name: "actions/checkout with v3.6.0",
			args: args{
				action: "uses: actions/checkout@v3.6.0",
			},
			want: &interfaces.EntityRef{
				Name:   "actions/checkout",
				Ref:    "f43a0e5ff2bd294095638e18286ca9a3d1956744",
				Type:   actions.ReferenceType,
				Tag:    "v3.6.0",
				Prefix: "uses: ",
			},
			wantErr: false,
		},
		{
			name: "actions/checkout with checksum returns checksum",
			args: args{
				action: "actions/checkout@1d96c772d19495a3b5c517cd2bc0cb401ea0529f",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "aquasecurity/trivy-action with 0.14.0",
			args: args{
				action: "aquasecurity/trivy-action@0.14.0",
			},
			want: &interfaces.EntityRef{
				Name:   "aquasecurity/trivy-action",
				Ref:    "2b6a709cf9c4025c5438138008beaddbb02086f0",
				Type:   actions.ReferenceType,
				Tag:    "0.14.0",
				Prefix: "",
			},
			wantErr: false,
		},
		{
			name: "aquasecurity/trivy-action with branch returns checksum",
			args: args{
				action: "aquasecurity/trivy-action@bump-trivy",
			},
			want: &interfaces.EntityRef{
				Name:   "aquasecurity/trivy-action",
				Ref:    "fb5e1b36be448e92ca98648c661bd7e9da1f1317",
				Type:   actions.ReferenceType,
				Tag:    "bump-trivy",
				Prefix: "",
			},
			wantErr: false,
		},
		{
			name: "aquasecurity/trivy-action with ignored branch returns error",
			args: args{
				action: "aquasecurity/trivy-action@main",
			},
			wantErr: true,
		},
		{
			name: "actions/checkout with invalid tag returns error",
			args: args{
				action: "actions/checkout@v4.1.1.1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "actions/checkout with invalid action returns error",
			args: args{
				action: "invalid-action@v4.1.1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "actions/checkout with empty action returns error",
			args: args{
				action: "@v4.1.1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "actions/checkout with empty tag returns error",
			args: args{
				action: "actions/checkout",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "actions/setup-node with v1 is an array",
			args: args{
				action: "actions/setup-node@v1",
			},
			want: &interfaces.EntityRef{
				Name:   "actions/setup-node",
				Ref:    "f1f314fca9dfce2769ece7d933488f076716723e",
				Type:   actions.ReferenceType,
				Tag:    "v1",
				Prefix: "",
			},
		},
		{
			name: "anchore/sbom-action/download-syft with a sub-action works",
			args: args{
				action: "anchore/sbom-action/download-syft@v0.14.3",
			},
			want: &interfaces.EntityRef{
				Name:   "anchore/sbom-action/download-syft",
				Ref:    "78fc58e266e87a38d4194b2137a3d4e9bcaf7ca1",
				Type:   actions.ReferenceType,
				Tag:    "v0.14.3",
				Prefix: "",
			},
		},
		{
			name: "invalid action reference",
			args: args{
				action: "invalid-reference",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing action tag",
			args: args{
				action: "actions/checkout",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "action with special characters",
			args: args{
				action: "actions/checkout@#$$%%^&*",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			conf := &config.Config{
				GHActions: config.GHActions{
					Filter: config.Filter{
						ExcludeBranches: []string{"main"},
					},
				},
				Images: config.Images{
					ImageFilter: config.ImageFilter{
						ExcludeTags: []string{"latest"},
					},
				},
			}
			r := NewGitHubActionsReplacer(conf).WithGitHubClientFromToken(os.Getenv("GITHUB_TOKEN"))
			got, err := r.ParseString(ctx, tt.args.action)
			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestReplacer_ParseContainerImagesInFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		before   string
		expected string
		modified bool
		wantErr  bool
	}{
		{
			name: "Replace image reference",
			before: `
version: v1
services:
  - name: kube-apiserver
    image: registry.k8s.io/kube-apiserver:v1.20.0
  - name: kube-controller-manager
    image: registry.k8s.io/kube-controller-manager:v1.15.0
  - name: minder-app
    image: minder:latest
`,
			expected: `
version: v1
services:
  - name: kube-apiserver
    image: registry.k8s.io/kube-apiserver@sha256:8b8125d7a6e4225b08f04f65ca947b27d0cc86380bf09fab890cc80408230114 # v1.20.0
  - name: kube-controller-manager
    image: registry.k8s.io/kube-controller-manager@sha256:835f32a5cdb30e86f35675dd91f9c7df01d48359ab8b51c1df866a2c7ea2e870 # v1.15.0
  - name: minder-app
    image: minder:latest
`,
			modified: true,
		},
		{
			name: "No image reference modification",
			before: `
version: v1
services:
  - name: minder-app
    image: minder:latest
`,
			expected: `
version: v1
services:
  - name: minder-app
    image: minder:latest
`,
			modified: false,
		},
		{
			name: "Invalid image reference format",
			before: `
version: v1
services:
  - name: invalid-service
    image: invalid@@reference
`,
			expected: `
version: v1
services:
  - name: invalid-service
    image: invalid@@reference
`,
			modified: false,
			wantErr:  false,
		},
		{
			name: "Multiple valid image references with one commented",
			before: `
version: v1
services:
  - name: kube-apiserver
    image: registry.k8s.io/kube-apiserver:v1.20.0
  - name: kube-controller-manager
    image: registry.k8s.io/kube-controller-manager:v1.15.0
  - name: minder-app
    image: minder:latest
  # - name: nginx
  #  image: nginx:latest
`,
			expected: `
version: v1
services:
  - name: kube-apiserver
    image: registry.k8s.io/kube-apiserver@sha256:8b8125d7a6e4225b08f04f65ca947b27d0cc86380bf09fab890cc80408230114 # v1.20.0
  - name: kube-controller-manager
    image: registry.k8s.io/kube-controller-manager@sha256:835f32a5cdb30e86f35675dd91f9c7df01d48359ab8b51c1df866a2c7ea2e870 # v1.15.0
  - name: minder-app
    image: minder:latest
  # - name: nginx
  #  image: nginx:latest
`,
			modified: true,
		},
		{
			name: "Valid image reference without specifying the tag",
			before: `
apiVersion: v1
kind: Pod
metadata:
  name: mount-host
  namespace: playground
spec:
  containers:
  - name: mount-host
    image: alpine
    command: ["sleep"]
    args: ["infinity"]
    volumeMounts:
    - name: host-root
      mountPath: /host
      readOnly: true
  volumes:
  - name: host-root
    hostPath:
      path: /
      type: Directory
`,
			modified: false,
		},
		{
			name: "A complex dockerfile",
			before: `
ARG BASE_IMAGE=alpine

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.2.1@sha256:8879a398dedf0aadaacfbd332b29ff2f84bc39ae6d4e9c0a1109db27ac5ba012 AS xx

FROM --platform=$BUILDPLATFORM golang:1.20.4-alpine3.16 AS builder

COPY --from=xx / /

RUN apk add --update alpine-sdk ca-certificates openssl clang lld

ARG TARGETPLATFORM

RUN xx-apk --update add musl-dev gcc

# lld has issues building static binaries for ppc so prefer ld for it
RUN [ "$(xx-info arch)" != "ppc64le" ] || XX_CC_PREFER_LINKER=ld xx-clang --setup-target-triple

RUN xx-go --wrap

WORKDIR /usr/local/src/dex

ARG GOPROXY

ENV CGO_ENABLED=1

COPY go.mod go.sum ./
COPY api/v2/go.mod api/v2/go.sum ./api/v2/
RUN go mod download

COPY . .

RUN make release-binary
RUN xx-verify /go/bin/dex && xx-verify /go/bin/docker-entrypoint

FROM alpine:3.18.2 AS stager

RUN mkdir -p /var/dex
RUN mkdir -p /etc/dex
COPY config.docker.yaml /etc/dex/

FROM alpine:3.18.2 AS gomplate

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

ENV GOMPLATE_VERSION=v3.11.4

RUN wget -O /usr/local/bin/gomplate \
  "https://github.com/hairyhenderson/gomplate/releases/download/${GOMPLATE_VERSION}/gomplate_${TARGETOS:-linux}-${TARGETARCH:-amd64}${TARGETVARIANT}" \
  && chmod +x /usr/local/bin/gomplate

# For Dependabot to detect base image versions
FROM alpine:3.18.2 AS alpine
FROM gcr.io/distroless/static:latest AS distroless

FROM $BASE_IMAGE

# Dex connectors, such as GitHub and Google logins require root certificates.
# Proper installations should manage those certificates, but it's a bad user
# experience when this doesn't work out of the box.
#
# See https://go.dev/src/crypto/x509/root_linux.go for Go root CA bundle locations.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=stager --chown=1001:1001 /var/dex /var/dex
COPY --from=stager --chown=1001:1001 /etc/dex /etc/dex

# Copy module files for CVE scanning / dependency analysis.
COPY --from=builder /usr/local/src/dex/go.mod /usr/local/src/dex/go.sum /usr/local/src/dex/
COPY --from=builder /usr/local/src/dex/api/v2/go.mod /usr/local/src/dex/api/v2/go.sum /usr/local/src/dex/api/v2/

COPY --from=builder /go/bin/dex /usr/local/bin/dex
COPY --from=builder /go/bin/docker-entrypoint /usr/local/bin/docker-entrypoint
COPY --from=builder /usr/local/src/dex/web /srv/dex/web

COPY --from=gomplate /usr/local/bin/gomplate /usr/local/bin/gomplate

USER 1001:1001

ENTRYPOINT ["/usr/local/bin/docker-entrypoint"]
CMD ["dex", "serve", "/etc/dex/config.docker.yaml"]
`,
			expected: `
ARG BASE_IMAGE=alpine

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.2.1@sha256:8879a398dedf0aadaacfbd332b29ff2f84bc39ae6d4e9c0a1109db27ac5ba012 AS xx

FROM --platform=$BUILDPLATFORM index.docker.io/library/golang:1.20.4-alpine3.16@sha256:6469405d7297f82d56195c90a3270b0806ef4bd897aa0628477d9959ab97a577 AS builder

COPY --from=xx / /

RUN apk add --update alpine-sdk ca-certificates openssl clang lld

ARG TARGETPLATFORM

RUN xx-apk --update add musl-dev gcc

# lld has issues building static binaries for ppc so prefer ld for it
RUN [ "$(xx-info arch)" != "ppc64le" ] || XX_CC_PREFER_LINKER=ld xx-clang --setup-target-triple

RUN xx-go --wrap

WORKDIR /usr/local/src/dex

ARG GOPROXY

ENV CGO_ENABLED=1

COPY go.mod go.sum ./
COPY api/v2/go.mod api/v2/go.sum ./api/v2/
RUN go mod download

COPY . .

RUN make release-binary
RUN xx-verify /go/bin/dex && xx-verify /go/bin/docker-entrypoint

FROM index.docker.io/library/alpine:3.18.2@sha256:82d1e9d7ed48a7523bdebc18cf6290bdb97b82302a8a9c27d4fe885949ea94d1 AS stager

RUN mkdir -p /var/dex
RUN mkdir -p /etc/dex
COPY config.docker.yaml /etc/dex/

FROM index.docker.io/library/alpine:3.18.2@sha256:82d1e9d7ed48a7523bdebc18cf6290bdb97b82302a8a9c27d4fe885949ea94d1 AS gomplate

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

ENV GOMPLATE_VERSION=v3.11.4

RUN wget -O /usr/local/bin/gomplate \
  "https://github.com/hairyhenderson/gomplate/releases/download/${GOMPLATE_VERSION}/gomplate_${TARGETOS:-linux}-${TARGETARCH:-amd64}${TARGETVARIANT}" \
  && chmod +x /usr/local/bin/gomplate

# For Dependabot to detect base image versions
FROM index.docker.io/library/alpine:3.18.2@sha256:82d1e9d7ed48a7523bdebc18cf6290bdb97b82302a8a9c27d4fe885949ea94d1 AS alpine
FROM gcr.io/distroless/static:latest AS distroless

FROM $BASE_IMAGE

# Dex connectors, such as GitHub and Google logins require root certificates.
# Proper installations should manage those certificates, but it's a bad user
# experience when this doesn't work out of the box.
#
# See https://go.dev/src/crypto/x509/root_linux.go for Go root CA bundle locations.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=stager --chown=1001:1001 /var/dex /var/dex
COPY --from=stager --chown=1001:1001 /etc/dex /etc/dex

# Copy module files for CVE scanning / dependency analysis.
COPY --from=builder /usr/local/src/dex/go.mod /usr/local/src/dex/go.sum /usr/local/src/dex/
COPY --from=builder /usr/local/src/dex/api/v2/go.mod /usr/local/src/dex/api/v2/go.sum /usr/local/src/dex/api/v2/

COPY --from=builder /go/bin/dex /usr/local/bin/dex
COPY --from=builder /go/bin/docker-entrypoint /usr/local/bin/docker-entrypoint
COPY --from=builder /usr/local/src/dex/web /srv/dex/web

COPY --from=gomplate /usr/local/bin/gomplate /usr/local/bin/gomplate

USER 1001:1001

ENTRYPOINT ["/usr/local/bin/docker-entrypoint"]
CMD ["dex", "serve", "/etc/dex/config.docker.yaml"]
`,
			modified: true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			r := NewContainerImagesReplacer(&config.Config{
				Images: config.Images{
					ImageFilter: config.ImageFilter{
						ExcludeTags: []string{"latest"},
					},
				},
			})
			modified, newContent, err := r.ParseFile(ctx, strings.NewReader(tt.before))
			if tt.wantErr {
				require.False(t, modified)
				require.Equal(t, tt.before, newContent)
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.modified {
				require.True(t, modified)
				if tt.expected != "" {
					require.Equal(t, tt.expected, newContent)
				} else {
					require.NotEmpty(t, tt.before, newContent)
				}
			} else {
				require.False(t, modified)
				require.Equal(t, tt.before, newContent)
			}

		})
	}
}

func TestReplacer_ParseGitHubActionsInFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		before         string
		expected       string
		regex          string
		modified       bool
		wantErr        bool
		useCustomRegex bool
	}{
		{
			name: "Replace image reference",
			before: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./minder/server.yml # this should not be replaced
      - uses: actions/checkout@v2
      - uses: xt0rted/markdownlint-problem-matcher@v1
      - name: "Run Markdown linter"
        uses: docker://avtodev/markdown-lint:v1
        with:
          args: src/*.md
`,
			expected: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./minder/server.yml # this should not be replaced
      - uses: actions/checkout@ee0669bd1cc54295c223e0bb666b733df41de1c5 # v2
      - uses: xt0rted/markdownlint-problem-matcher@c17ca40d1376f60aba7e7d38a8674a3f22f7f5b0 # v1
      - name: "Run Markdown linter"
        uses: docker://index.docker.io/avtodev/markdown-lint@sha256:6aeedc2f49138ce7a1cd0adffc1b1c0321b841dc2102408967d9301c031949ee # v1
        with:
          args: src/*.md
`,
			modified: true,
			wantErr:  false,
		},
		{
			name: "Replace actions with tags, not with branches",
			before: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./minder/server.yml # this should not be replaced
      - uses: actions/checkout@v2
      - uses: aquasecurity/trivy-action@main
      - name: "Run Markdown linter"
        uses: docker://avtodev/markdown-lint:v1
        with:
          args: src/*.md
`,
			expected: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./minder/server.yml # this should not be replaced
      - uses: actions/checkout@ee0669bd1cc54295c223e0bb666b733df41de1c5 # v2
      - uses: aquasecurity/trivy-action@main
      - name: "Run Markdown linter"
        uses: docker://index.docker.io/avtodev/markdown-lint@sha256:6aeedc2f49138ce7a1cd0adffc1b1c0321b841dc2102408967d9301c031949ee # v1
        with:
          args: src/*.md
`,
			modified: true,
			wantErr:  false,
		},
		{
			name: "No action reference modification",
			before: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./minder/server.yml # this should not be replaced
      # - uses: actions/checkout@v2
`,
			expected: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./minder/server.yml # this should not be replaced
      # - uses: actions/checkout@v2
`,
			modified: false,
		},
		{
			name: "Invalid action reference format",
			before: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: invalid@@reference
`,
			expected: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: invalid@@reference
`,
			modified: false,
			wantErr:  false,
		},
		{
			name: "Multiple valid action references",
			before: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./minder/server.yml # this should not be replaced
      - uses: actions/checkout@v2
      - uses: xt0rted/markdownlint-problem-matcher@v1
`,
			expected: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./minder/server.yml # this should not be replaced
      - uses: actions/checkout@ee0669bd1cc54295c223e0bb666b733df41de1c5 # v2
      - uses: xt0rted/markdownlint-problem-matcher@c17ca40d1376f60aba7e7d38a8674a3f22f7f5b0 # v1
`,
			modified: true,
		},
		{
			name: "Fail with custom regex",
			before: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./minder/server.yml # this should not be replaced
      - uses: actions/checkout@v2
      - uses: xt0rted/markdownlint-problem-matcher@v1
      - name: "Run Markdown linter"
        uses: docker://avtodev/markdown-lint:v1
        with:
          args: src/*.md
`,
			expected: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./minder/server.yml # this should not be replaced
      - uses: actions/checkout@v2
      - uses: xt0rted/markdownlint-problem-matcher@v1
      - name: "Run Markdown linter"
        uses: docker://avtodev/markdown-lint:v1
        with:
          args: src/*.md
`,
			modified:       false,
			wantErr:        false,
			regex:          "invalid-regexp",
			useCustomRegex: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			r := NewGitHubActionsReplacer(&config.Config{
				GHActions: config.GHActions{
					Filter: config.Filter{
						ExcludeBranches: []string{"*"},
					},
				},
			}).WithGitHubClientFromToken(os.Getenv(cli.GitHubTokenEnvKey))
			if tt.useCustomRegex {
				r = r.WithUserRegex(tt.regex)
			}
			modified, newContent, err := r.ParseFile(ctx, strings.NewReader(tt.before))
			if tt.modified {
				require.True(t, modified)
				require.Equal(t, tt.expected, newContent)
			} else {
				require.False(t, modified)
				require.Equal(t, tt.before, newContent)
			}
			if tt.wantErr {
				require.False(t, modified)
				require.Equal(t, tt.before, newContent)
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, newContent)
		})
	}
}

func TestReplacer_NewGitHubActionsReplacer(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	tests := []struct {
		name string
		cfg  *config.Config
	}{
		{name: "valid config", cfg: cfg},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := NewGitHubActionsReplacer(tt.cfg)
			require.NotNil(t, r)
			require.IsType(t, &Replacer{}, r)
			require.IsType(t, actions.New(), r.parser)
		})
	}
}

func TestReplacer_NewContainerImagesReplacer(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	tests := []struct {
		name string
		cfg  *config.Config
	}{
		{name: "valid config", cfg: cfg},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := NewContainerImagesReplacer(tt.cfg)
			require.NotNil(t, r)
			require.IsType(t, &Replacer{}, r)
			require.IsType(t, image.New(), r.parser)
		})
	}
}

func TestReplacer_WithGitHubClient(t *testing.T) {
	t.Parallel()

	r := &Replacer{}
	tests := []struct {
		name  string
		token string
	}{
		{name: "valid token", token: "valid_token"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r = r.WithGitHubClientFromToken(tt.token)
			require.NotNil(t, r)
			require.IsType(t, ghrest.NewClient(tt.token), r.rest)
		})
	}
}

func TestReplacer_WithUserRegex(t *testing.T) {
	t.Parallel()

	r := &Replacer{parser: actions.New()}
	tests := []struct {
		name  string
		regex string
	}{
		{name: "valid regex", regex: `^test-regex$`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r = r.WithUserRegex(tt.regex)
			require.Equal(t, tt.regex, r.parser.GetRegex())
		})
	}
}

func TestReplacer_WithCacheDisabled(t *testing.T) {
	t.Parallel()

	r := &Replacer{parser: actions.New()}
	tests := []struct {
		name string
	}{
		{name: "disable cache"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r = r.WithCacheDisabled()
			// we don't test if this passed here because it's an internal implementation detail
			// but let's ensure we don't panic for some reason
		})
	}
}

func TestReplacer_ParsePathInFS(t *testing.T) {
	t.Parallel()

	r := &Replacer{parser: actions.New(), cfg: config.Config{}}
	fs := memfs.New()
	tests := []struct {
		name    string
		base    string
		wantErr bool
	}{
		{name: "valid base", base: "some-base", wantErr: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := r.ParsePathInFS(context.Background(), fs, tt.base)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestReplacer_ListPathInFS(t *testing.T) {
	t.Parallel()

	r := &Replacer{parser: actions.New(), cfg: config.Config{}}
	fs := memfs.New()
	tests := []struct {
		name    string
		base    string
		wantErr bool
	}{
		{name: "valid base", base: "some-base", wantErr: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := r.ListPathInFS(fs, tt.base)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestReplacer_ListContainerImagesInFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		before         string
		expected       *ListResult
		regex          string
		wantErr        bool
		useCustomRegex bool
	}{
		{
			name: "Lust image reference",
			before: `
version: v1
services:
 - name: kube-apiserver
   image: registry.k8s.io/kube-apiserver:v1.20.0
 - name: kube-controller-manager
   image: registry.k8s.io/kube-controller-manager:v1.15.0
 - name: minder-app
   image: minder:latest
`,
			expected: &ListResult{
				Entities: []interfaces.EntityRef{
					{
						Name: "registry.k8s.io/kube-apiserver",
						Ref:  "v1.20.0",
						Type: image.ReferenceType,
					},
					{
						Name: "registry.k8s.io/kube-controller-manager",
						Ref:  "v1.15.0",
						Type: image.ReferenceType,
					},
					{
						Name: "minder",
						Ref:  "latest",
						Type: image.ReferenceType,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "No image reference modification",
			before: `
		version: v1
		services:
		- name: minder-app
		  # image: minder:latest
		`,
			expected: &ListResult{
				Entities: []interfaces.EntityRef{},
			},
			wantErr: false,
		},
		{
			name: "Invalid image reference format",
			before: `
		version: v1
		services:
		- name: invalid-service
		  image: invalid@@reference
		`,
			expected: &ListResult{
				Entities: []interfaces.EntityRef{},
			},
			wantErr: false,
		},
		{
			name: "Multiple valid image references with one commented",
			before: `
		version: v1
		services:
		- name: kube-apiserver
		  image: registry.k8s.io/kube-apiserver@sha256:8b8125d7a6e4225b08f04f65ca947b27d0cc86380bf09fab890cc80408230114 # v1.20.0
		- name: kube-controller-manager
		  image: registry.k8s.io/kube-controller-manager@sha256:835f32a5cdb30e86f35675dd91f9c7df01d48359ab8b51c1df866a2c7ea2e870 # v1.15.0
		- name: minder-app
		  image: minder:latest
		# - name: nginx
		#  image: nginx:latest
		`,
			expected: &ListResult{
				Entities: []interfaces.EntityRef{
					{
						Name: "registry.k8s.io/kube-apiserver",
						Ref:  "sha256:8b8125d7a6e4225b08f04f65ca947b27d0cc86380bf09fab890cc80408230114",
						Type: image.ReferenceType,
					},
					{
						Name: "registry.k8s.io/kube-controller-manager",
						Ref:  "sha256:835f32a5cdb30e86f35675dd91f9c7df01d48359ab8b51c1df866a2c7ea2e870",
						Type: image.ReferenceType,
					},
					{
						Name: "minder",
						Ref:  "latest",
						Type: image.ReferenceType,
					},
				},
			},
		},
		{
			name: "Valid image reference without specifying the tag",
			before: `
apiVersion: v1
kind: Pod
metadata:
  name: mount-host
  namespace: playground
spec:
  containers:
  - name: mount-host
    image: alpine
    command: ["sleep"]
    args: ["infinity"]
    volumeMounts:
    - name: host-root
      mountPath: /host
      readOnly: true
  volumes:
  - name: host-root
    hostPath:
      path: /
      type: Directory
		`,
			expected: &ListResult{
				Entities: []interfaces.EntityRef{
					{
						Name: "alpine",
						Ref:  "latest",
						Type: image.ReferenceType,
					},
				},
			},
		},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := NewContainerImagesReplacer(&config.Config{})
			listRes, err := r.ListInFile(strings.NewReader(tt.before))
			if tt.wantErr {
				require.Nil(t, listRes)
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, len(tt.expected.Entities), len(listRes.Entities))
			for _, entity := range tt.expected.Entities {
				require.Contains(t, listRes.Entities, entity)
			}
		})
	}
}

func TestReplacer_ListGitHubActionsInFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		before         string
		expected       *ListResult
		regex          string
		wantErr        bool
		useCustomRegex bool
	}{
		{
			name: "List image reference",
			before: `
name: Linter
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./minder/server.yml # this should not be listed
      - uses: actions/checkout@v2
      - uses: xt0rted/markdownlint-problem-matcher@v1
      - name: "Run Markdown linter"
        uses: docker://avtodev/markdown-lint:v1
        with:
          args: src/*.md
`,
			expected: &ListResult{
				Entities: []interfaces.EntityRef{
					{
						Name: "actions/checkout",
						Ref:  "v2",
						Type: actions.ReferenceType,
					},
					{
						Name: "xt0rted/markdownlint-problem-matcher",
						Ref:  "v1",
						Type: actions.ReferenceType,
					},
					{
						Name: "avtodev/markdown-lint",
						Ref:  "v1",
						Type: image.ReferenceType,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "No action references",
			before: `
		name: Linter
		on: pull_request
		jobs:
		 build:
		   runs-on: ubuntu-latest
		   steps:
		     - uses: ./minder/server.yml # this should not be replaced
		     # - uses: actions/checkout@v2
		`,
			expected: &ListResult{
				Entities: []interfaces.EntityRef{},
			},
			wantErr: false,
		},
		{
			name: "Invalid action reference format",
			before: `
		name: Linter
		on: pull_request
		jobs:
		 build:
		   runs-on: ubuntu-latest
		   steps:
		     - uses: invalid@@reference
		`,
			expected: &ListResult{
				Entities: []interfaces.EntityRef{},
			},
			wantErr: false,
		},
		{
			name: "Multiple valid action references",
			before: `
		name: Linter
		on: pull_request
		jobs:
		 build:
		   runs-on: ubuntu-latest
		   steps:
		     - uses: ./minder/server.yml # this should not be replaced
		     - uses: actions/checkout@ee0669bd1cc54295c223e0bb666b733df41de1c5 # v2
		     - uses: xt0rted/markdownlint-problem-matcher@c17ca40d1376f60aba7e7d38a8674a3f22f7f5b0 # v1
		`,
			expected: &ListResult{
				Entities: []interfaces.EntityRef{
					{
						Name: "actions/checkout",
						Ref:  "ee0669bd1cc54295c223e0bb666b733df41de1c5",
						Type: actions.ReferenceType,
					},
					{
						Name: "xt0rted/markdownlint-problem-matcher",
						Ref:  "c17ca40d1376f60aba7e7d38a8674a3f22f7f5b0",
						Type: actions.ReferenceType,
					},
				},
			},
		},
		{
			name: "Fail with custom regex",
			before: `
		name: Linter
		on: pull_request
		jobs:
		 build:
		   runs-on: ubuntu-latest
		   steps:
		     - uses: ./minder/server.yml # this should not be replaced
		     - uses: actions/checkout@v2
		     - uses: xt0rted/markdownlint-problem-matcher@v1
		     - name: "Run Markdown linter"
		       uses: docker://avtodev/markdown-lint:v1
		       with:
		         args: src/*.md
		`,
			expected: &ListResult{
				Entities: []interfaces.EntityRef{},
			},
			wantErr:        false,
			regex:          "invalid-regexp",
			useCustomRegex: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := NewGitHubActionsReplacer(&config.Config{}).WithGitHubClientFromToken(os.Getenv(cli.GitHubTokenEnvKey))
			if tt.useCustomRegex {
				r = r.WithUserRegex(tt.regex)
			}
			listRes, err := r.ListInFile(strings.NewReader(tt.before))
			if tt.wantErr {
				require.Nil(t, listRes)
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, len(tt.expected.Entities), len(listRes.Entities))
			for _, entity := range tt.expected.Entities {
				require.Contains(t, listRes.Entities, entity)
			}
		})
	}
}
