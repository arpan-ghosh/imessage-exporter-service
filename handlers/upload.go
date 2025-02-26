package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"imessage-exporter-webservice/utils"
)

// UploadChatDB handles file upload and stores in S3
func UploadChatDB(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB max file size
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save file locally before upload
	tempDir := os.TempDir()
	filePath := filepath.Join(tempDir, handler.Filename)
	outFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	// Upload to S3
	s3URL, err := utils.UploadToS3(filePath, handler.Filename)
	if err != nil {
		http.Error(w, "Failed to upload to S3", http.StatusInternalServerError)
		return
	}

	// Return S3 URL
	response := map[string]string{"message": "Upload successful", "s3_url": s3URL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
