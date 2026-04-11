package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	plugrpc "github.com/sailorhq/plug/sdk/proto"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap/zaptest"
	"github.com/sailorhq/sailor/internal/signal"
)

func TestPostProjectCreateSignalHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           []byte
		expectedStatus int
	}{
		{
			name:           "invalid json body",
			body:           []byte("{invalid json}"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "valid json body",
			body: func() []byte {
				req := plugrpc.ProjectCreateRequest{
					Ns:  "test-ns",
					App: "test-app",
				}
				b, _ := json.Marshal(req)
				return b
			}(),
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sc := &SailorCore{
				plugman: signal.NewPlugManager(zaptest.NewLogger(t)),
			}

			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetBody(tc.body)

			sc.PostProjectCreateSignalHandler(ctx)

			if ctx.Response.StatusCode() != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, ctx.Response.StatusCode())
			}
		})
	}
}

func TestPostDeploySignalHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           []byte
		expectedStatus int
	}{
		{
			name:           "invalid json body",
			body:           []byte("{invalid json}"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "valid json body",
			body: func() []byte {
				req := plugrpc.DeployRequest{
					Ns:  "test-ns",
					App: "test-app",
				}
				b, _ := json.Marshal(req)
				return b
			}(),
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sc := &SailorCore{
				plugman: signal.NewPlugManager(zaptest.NewLogger(t)),
			}

			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetBody(tc.body)

			sc.PostDeploySignalHandler(ctx)

			if ctx.Response.StatusCode() != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, ctx.Response.StatusCode())
			}
		})
	}
}
