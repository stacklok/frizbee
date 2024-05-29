package ghrest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

const ContextTimeout = 4 * time.Second

// nolint:gocyclo
func TestClientFunctions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		token          string
		method         string
		url            string
		mockResponse   *gock.Response
		expectedMethod string
		expectedURL    string
		expectError    bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "NewClient",
			token:          "test_token",
			expectedMethod: "",
			expectedURL:    "",
		},
		{
			name:           "NewRequest GET",
			token:          "",
			method:         "GET",
			url:            "test_url",
			expectedMethod: http.MethodGet,
			expectedURL:    "https://api.github.com/test_url",
		},
		{
			name:           "Do successful request",
			token:          "",
			method:         "GET",
			url:            "test",
			mockResponse:   gock.New("https://api.github.com").Get("/test").Reply(200).BodyString(`{"message": "hello world"}`),
			expectedMethod: http.MethodGet,
			expectedURL:    "https://api.github.com/test",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message": "hello world"}`,
		},
		{
			name:           "Do failed request",
			token:          "",
			method:         "GET",
			url:            "test",
			mockResponse:   gock.New("https://api.github.com").Get("/test").ReplyError(errors.New("failed request")),
			expectedMethod: http.MethodGet,
			expectedURL:    "https://api.github.com/test",
			expectError:    true,
		},
	}

	for _, tt := range testCases {
		tt := tt

		if tt.mockResponse != nil {
			defer gock.Off()
			//gock.DisableNetworking()
			//t.Logf("Mock response configured for %s %s", tt.method, tt.url)
		}

		client := NewClient(tt.token)

		if tt.name == "NewClient" {
			assert.NotNil(t, client, "NewClient returned nil")
			assert.NotNil(t, client.client, "NewClient returned client with nil GitHub client")
			return
		}

		req, err := client.NewRequest(tt.method, tt.url, nil)
		require.NoError(t, err)
		require.Equal(t, req.Method, tt.expectedMethod)
		require.Equal(t, req.URL.String(), tt.expectedURL)

		if tt.name == "NewRequest GET" {
			return
		}

		ctx := context.Background()

		resp, err := client.Do(ctx, req)
		if tt.expectError {
			require.NotNil(t, err, "Expected error, got nil")
			require.Nil(t, resp, "Expected nil response, got %v", resp)
			return
		}
		require.Nil(t, err, "Expected no error, got %v", err)

		require.Equal(t, resp.StatusCode, tt.expectedStatus)

		body, err := io.ReadAll(resp.Body)
		require.Nil(t, err)
		require.Equal(t, string(body), tt.expectedBody)
		defer resp.Body.Close() // nolint:errcheck
	}
}
