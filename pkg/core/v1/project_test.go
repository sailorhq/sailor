package v1_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/codekidx/sailor/pkg/core/v1"
)

// TODO :: plugin fasthttp server for testing
func TestCreateProject(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		app            string
		token          string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    bool
		expectedStatus int
	}{
		{
			name:      "successful request",
			namespace: "sailor",
			app:       "backend-core",
			token:     "valid-token-123",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "PUT" {
					t.Errorf("Expected PUT method, got %s", r.Method)
				}

				// Verify URL path
				expectedPath := "/api/v1/project/sailor/backend-core"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Verify headers
				if r.Header.Get("x-token") != "valid-token-123" {
					t.Errorf("Expected x-token header to be 'valid-token-123', got '%s'", r.Header.Get("x-token"))
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type header to be 'application/json', got '%s'", r.Header.Get("Content-Type"))
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "success"}`))
			},
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:      "server error",
			namespace: "sailor",
			app:       "backend-core",
			token:     "valid-token-123",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal server error"}`))
			},
			expectError:    false, // HTTP errors don't cause Go errors, just non-2xx status
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "different namespace and app",
			namespace: "test-ns",
			app:       "test-app",
			token:     "test-token",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/v1/project/test-ns/test-app"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "ok"}`))
			},
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			// Create client with test server URL
			client := v1.CoreV1(server.URL)

			// Make the request
			resp, err := client.CreateProject(tt.namespace, tt.app, tt.token)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// If we got a response, check the status code
			if resp != nil {
				if resp.Key != fmt.Sprintf("%s-%s", tt.namespace, tt.app) {
					t.Errorf("Expected project key %s, got %s", fmt.Sprintf("%s-%s", tt.namespace, tt.app), resp.Key)
				}
			}
		})
	}
}
