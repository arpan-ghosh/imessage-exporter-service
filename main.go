package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"imessage-exporter-webservice/config"
	"imessage-exporter-webservice/routes"
)

func main() {
	fmt.Println("✅ Starting iMessage Exporter API...")
	config.InitAWS()

	r := routes.SetupRoutes()

	fmt.Println("✅ Server started on :8080")

	r.Use(authMiddleware)

	log.Fatal(http.ListenAndServe(":8080", r))

}

// Middleware to check API Key
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := os.Getenv("API_KEY")
		reqKey := r.Header.Get("X-API-KEY")

		if reqKey != apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}