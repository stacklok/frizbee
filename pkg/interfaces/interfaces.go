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

package interfaces

import (
	"context"
	"github.com/stacklok/frizbee/internal/store"
	"github.com/stacklok/frizbee/pkg/config"
	"net/http"
)

// EntityRef represents an action reference.
type EntityRef struct {
	Name string `json:"name"`
	Ref  string `json:"ref"`
	Type string `json:"type"`
}

type Parser interface {
	GetRegex() string
	Replace(ctx context.Context, matchedLine string, restIf REST, cfg config.Config, cache store.RefCacher, keepPrefix bool) (string, error)
	ConvertToEntityRef(reference string) (*EntityRef, error)
}

// The REST interface allows to wrap clients to talk to remotes
// When talking to GitHub, wrap a github client to provide this interface
type REST interface {
	// NewRequest creates an HTTP request.
	NewRequest(method, url string, body any) (*http.Request, error)
	// Do executes an HTTP request.
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}
