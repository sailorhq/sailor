package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/valyala/fasthttp"
)

func TestAuthOIDCHandler(t *testing.T) {
	// happy path tests might need an actual OIDC provider to get the provider URL.
	// We can test at least the cases where the handler fails to get setting,
	// missing OIDC config, and no fp query param.

	tests := []struct {
		name           string
		setupMock      func(*MockSail)
		queryArgs      string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "missing fp query arg",
			setupMock: func(m *MockSail) {},
			queryArgs: "",
			expectedStatus: http.StatusInternalServerError,
			expectedMsg: "unknown caller",
		},
		{
			name: "failed to get sailor setting",
			setupMock: func(m *MockSail) {
				m.GetSailorSettingFunc = func() (*v1.SailorSetting, error) {
					return nil, errors.New("db error")
				}
			},
			queryArgs: "fp=123",
			expectedStatus: http.StatusInternalServerError,
			expectedMsg: "db error",
		},
		{
			name: "missing OIDC setting",
			setupMock: func(m *MockSail) {
				m.GetSailorSettingFunc = func() (*v1.SailorSetting, error) {
					return &v1.SailorSetting{}, nil
				}
			},
			queryArgs: "fp=123",
			expectedStatus: http.StatusNotFound,
			expectedMsg: "oidc sailor settings not present",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSail := &MockSail{}
			tc.setupMock(mockSail)

			sc := &SailorCore{
				SailorSail: mockSail,
			}

			ctx := &fasthttp.RequestCtx{}
			ctx.Request.URI().SetQueryString(tc.queryArgs)

			sc.AuthOIDCHandler(ctx)

			if ctx.Response.StatusCode() != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, ctx.Response.StatusCode())
			}

			var resp ResponseMessage
			err := json.Unmarshal(ctx.Response.Body(), &resp)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if resp.Message != tc.expectedMsg {
				t.Errorf("expected message '%s', got '%s'", tc.expectedMsg, resp.Message)
			}
		})
	}
}
