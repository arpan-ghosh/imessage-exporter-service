package handlers

import (
	"encoding/json"
	"net/http"
)

// JSONResponse sends a JSON response
func JSONResponse(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"message": message,
		"data":    data,
	}

	json.NewEncoder(w).Encode(response)
}
