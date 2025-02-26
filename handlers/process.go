package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"imessage-exporter-webservice/utils"
)

// RequestBody defines the JSON structure expected in /process
type RequestBody struct {
	S3URL        string `json:"s3_url"`
	PhoneNumber  string `json:"phone_number"`
	ContactName  string `json:"contact_name"`
}

// ProcessChatDB downloads the file, runs imessage-exporter, and uploads results
func ProcessChatDB(w http.ResponseWriter, r *http.Request) {
	var reqBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Download file from S3
	tempDir := os.TempDir()
	localChatDB := filepath.Join(tempDir, "chat.db")
	err = utils.DownloadFromS3(reqBody.S3URL, localChatDB)
	if err != nil {
		http.Error(w, "Failed to download chat.db", http.StatusInternalServerError)
		return
	}

	// Run imessage-exporter
	outputDir := filepath.Join(tempDir, "output")
	err = utils.RunExporter(localChatDB, outputDir, reqBody.PhoneNumber, reqBody.ContactName)
	if err != nil {
		http.Error(w, "Failed to process chat.db", http.StatusInternalServerError)
		return
	}

	// Upload results to S3
	exportedFiles, err := utils.UploadFolderToS3(outputDir)
	if err != nil {
		http.Error(w, "Failed to upload extracted messages", http.StatusInternalServerError)
		return
	}

	// Return result URLs
	response := map[string]interface{}{
		"message": "Processing complete",
		"files":   exportedFiles,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
