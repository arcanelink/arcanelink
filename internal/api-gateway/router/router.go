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
	api.HandleFunc("/auth/logout", apiHandler.Logout).Methods("POST")

	// Message routes (RESTful)
	api.HandleFunc("/messages", apiHandler.SendDirect).Methods("POST")                    // Send direct message
	api.HandleFunc("/messages", apiHandler.GetDirectHistory).Methods("GET")               // Get direct message history
	api.HandleFunc("/rooms/{room_id}/messages", apiHandler.SendRoomMessage).Methods("POST") // Send room message
	api.HandleFunc("/rooms/{room_id}/messages", apiHandler.GetRoomHistory).Methods("GET")   // Get room message history

	// Room routes (RESTful)
	api.HandleFunc("/rooms", apiHandler.CreateRoom).Methods("POST")                          // Create room
	api.HandleFunc("/rooms", apiHandler.GetRooms).Methods("GET")                             // Get user's rooms
	api.HandleFunc("/rooms/{room_id}", apiHandler.GetRoomState).Methods("GET")               // Get room state
	api.HandleFunc("/rooms/{room_id}", apiHandler.DeleteRoom).Methods("DELETE")              // Delete room
	api.HandleFunc("/rooms/{room_id}/members", apiHandler.GetRoomMembers).Methods("GET")     // Get room members
	api.HandleFunc("/rooms/{room_id}/members", apiHandler.ManageRoomMember).Methods("POST", "DELETE") // Join/Leave/Invite

	// Sync route
	api.HandleFunc("/sync", apiHandler.Sync).Methods("GET")

	// File routes
	api.HandleFunc("/files", apiHandler.UploadFile).Methods("POST")                      // Upload file
	api.HandleFunc("/files/{file_id}", apiHandler.DownloadFile).Methods("GET")           // Download file
	api.HandleFunc("/files/{file_id}/info", apiHandler.GetFileInfo).Methods("GET")       // Get file info

	// Link preview
	api.HandleFunc("/link_preview", apiHandler.GetLinkPreview).Methods("GET")

	// Debug: print registered routes
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		if len(methods) > 0 {
			println("Route:", path, "Methods:", methods[0])
		}
		return nil
	})

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
