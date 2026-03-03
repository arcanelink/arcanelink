package middleware

import (
	"context"
	"net/http"
	"strings"

	pb "github.com/arcane/arcanelink/pkg/proto/auth"
	"google.golang.org/grpc"
)

type AuthMiddleware struct {
	authClient pb.AuthServiceClient
}

func NewAuthMiddleware(authServiceAddr string) (*AuthMiddleware, error) {
	conn, err := grpc.Dial(authServiceAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewAuthServiceClient(conn)
	return &AuthMiddleware{authClient: client}, nil
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"UNAUTHORIZED","message":"Missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"UNAUTHORIZED","message":"Invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Validate token with auth service
		resp, err := m.authClient.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
			Token: token,
		})
		if err != nil || !resp.Valid {
			http.Error(w, `{"error":"UNAUTHORIZED","message":"Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Add user ID to context
		ctx := context.WithValue(r.Context(), "user_id", resp.UserId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts user ID from request context
func GetUserID(r *http.Request) string {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		return ""
	}
	return userID
}
