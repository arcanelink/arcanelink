package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/arcane/arcanelink/internal/api-gateway/handler"
	"github.com/arcane/arcanelink/internal/api-gateway/middleware"
)

func SetupRouter(apiHandler *handler.APIHandler, authMiddleware *middleware.AuthMiddleware, rateLimiter *middleware.RateLimiter) *mux.Router {
	r := mux.NewRouter()

	// CORS middleware
	r.Use(corsMiddleware)

	// Public routes (no authentication required)
	r.HandleFunc("/_api/v1/auth/login", apiHandler.Login).Methods("POST")
	r.HandleFunc("/_api/v1/auth/register", apiHandler.Register).Methods("POST")

	// Protected routes (authentication required)
	api := r.PathPrefix("/_api/v1").Subrouter()
	api.Use(authMiddleware.Authenticate)
	api.Use(rateLimiter.Limit(100)) // 100 requests per second per user

	// Auth routes
	api.HandleFunc("/logout", apiHandler.Logout).Methods("POST")

	// Message routes
	api.HandleFunc("/send_direct", apiHandler.SendDirect).Methods("POST")
	api.HandleFunc("/sync", apiHandler.Sync).Methods("GET")
	api.HandleFunc("/direct_history", apiHandler.GetDirectHistory).Methods("GET")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
