package webserver

import (
	"io"
	"log/slog"
	"myproject/application"
	"myproject/domain"
	"myproject/infrastructure/testhelpers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_Authenticate(t *testing.T) {
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantUserID int
		expectCall bool
	}{
		{"no header", "", http.StatusUnauthorized, 0, false},
		{"invalid format", "Bearer", http.StatusUnauthorized, 0, false},
		{"invalid token", "Bearer invalid-token", http.StatusUnauthorized, 0, false},
		{"valid token", "Bearer valid-jwt", http.StatusOK, 123, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			stubTokenGenerator := &testhelpers.StubTokenGenerator{
				Token: "valid-jwt",
				Claims: &domain.Claims{
					UserID: tc.wantUserID,
				},
				Err: nil,
			}

			// Для случая с invalid token
			if tc.name == "invalid token" {
				stubTokenGenerator.Err = assert.AnError
			}

			middleware := NewAuthMiddleware(stubTokenGenerator, testLogger)

			var capturedUserID int
			handler := middleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
				userID, err := application.GetUserIDFromContext(r.Context())
				if err == nil {
					capturedUserID = userID
				}
			})

			// Act
			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			req.Header.Set("Authorization", tc.authHeader)
			rr := httptest.NewRecorder()
			handler(rr, req)

			// Assert
			assert.Equal(t, tc.wantStatus, rr.Code)

			if tc.expectCall {
				assert.Equal(t, tc.wantUserID, capturedUserID)
			}
		})
	}
}
