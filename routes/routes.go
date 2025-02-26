package routes

import (
	"imessage-exporter-webservice/handlers"

	"github.com/gorilla/mux"
)

// SetupRoutes initializes API routes
func SetupRoutes() *mux.Router {
	r := mux.NewRouter()
	
	// Existing routes
	r.HandleFunc("/upload", handlers.UploadChatDB).Methods("POST")

	// NEW: Add the processing endpoint
	r.HandleFunc("/process", handlers.ProcessChatDB).Methods("POST")

	return r
}
