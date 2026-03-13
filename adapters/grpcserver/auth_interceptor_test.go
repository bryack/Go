package grpcserver

import (
	"context"
	"io"
	"log/slog"
	"myproject/application"
	"myproject/domain"
	"myproject/infrastructure/testhelpers"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryInterceptor(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		ctx         context.Context
		tokenErr    error
		wantError   bool
		wantErrCode codes.Code
		wantUserID  int
	}{
		{
			name:      "public method - no token required",
			method:    "/grpcserver.TaskManager/Register",
			ctx:       context.Background(),
			wantError: false,
		},
		{
			name:        "missing metadata",
			method:      "/grpcserver.TaskManager/CreateTask",
			ctx:         context.Background(),
			wantError:   true,
			wantErrCode: codes.Unauthenticated,
		},
		{
			name:       "valid token",
			method:     "/grpcserver.TaskManager/CreateTask",
			ctx:        createCtxWithToken("valid-jwt"),
			tokenErr:   nil,
			wantError:  false,
			wantUserID: 123,
		},
		{
			name:        "invalid token",
			method:      "/grpcserver.TaskManager/CreateTask",
			ctx:         createCtxWithToken("invalid-jwt"),
			tokenErr:    assert.AnError,
			wantError:   true,
			wantErrCode: codes.Unauthenticated,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenGenerator := &testhelpers.StubTokenGenerator{
				Token:  "test-token",
				Claims: &domain.Claims{UserID: tt.wantUserID},
				Err:    tt.tokenErr,
			}

			interceptor := NewAuthInterceptor(tokenGenerator, logger)

			handler := func(ctx context.Context, req any) (any, error) {
				if !tt.wantError && tt.wantUserID != 0 {
					userID, err := application.GetUserIDFromContext(ctx)
					assert.NoError(t, err)
					assert.Equal(t, tt.wantUserID, userID)
				}
				return &CreateTaskReply{TaskId: 1}, nil
			}

			request := &CreateTaskRequest{Description: "test"}
			info := &grpc.UnaryServerInfo{FullMethod: tt.method}

			resp, err := interceptor.UnaryInterceptor(tt.ctx, request, info, handler)
			if tt.wantError {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErrCode, status.Code(err))
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func createCtxWithToken(token string) context.Context {
	md := metadata.Pairs("authorization", "Bearer "+token)
	return metadata.NewIncomingContext(context.Background(), md)
}
