package main

import (
	"fmt"
	"log"
	"net/http"

	"imessage-exporter-webservice/config"
	"imessage-exporter-webservice/handlers"

	"github.com/gorilla/mux"
)

func main() {
	config.InitAWS()

	r := mux.NewRouter()

	// Define API routes
	r.HandleFunc("/upload", handlers.UploadChatDB).Methods("POST")
	r.HandleFunc("/process", handlers.ProcessChatDB).Methods("POST")

	fmt.Println("✅ Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
