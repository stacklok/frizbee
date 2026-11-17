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

// Package ghrest provides a wrapper around the go-github client that implements the internal REST API
package ghrest

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/google/go-github/v93/github"
)

// Client is the struct that contains the GitHub REST API client
// this struct implements the REST API
type Client struct {
	client *github.Client
}

// NewClient creates a new instance of GhRest
func NewClient(token string) *Client {
	var opts []github.ClientOptionsFunc
	if token != "" {
		opts = append(opts, github.WithAuthToken(token))
	}

	// NewClient only fails on invalid options. The only option we pass is a
	// non-empty token, so an error here is a programming bug, not a runtime
	// condition callers could handle.
	ghcli, err := github.NewClient(opts...)
	if err != nil {
		panic("ghrest: creating GitHub client: " + err.Error())
	}
	return &Client{
		client: ghcli,
	}
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// which will be resolved to the BaseURL of the Client. Relative URLS should
// always be specified without a preceding slash. If specified, the value
// pointed to by body is JSON encoded and included as the request body.
func (c *Client) NewRequest(method, requestUrl string, body any) (*http.Request, error) {
	// The context is attached later in Do, so callers keep a single place to
	// pass it and cancellation still applies to the actual round trip.
	return c.client.NewRequest(context.Background(), method, requestUrl, body)
}

// Do sends an API request and returns the API response.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var buf bytes.Buffer

	// The GitHub client closes the response body, so we need to capture it
	// in a buffer so that we can return it to the caller
	resp, err := c.client.Do(req.WithContext(ctx), &buf)
	if err != nil && resp == nil {
		return nil, err
	}

	if resp.Response != nil {
		resp.Body = io.NopCloser(&buf)
	}

	return resp.Response, err
}
