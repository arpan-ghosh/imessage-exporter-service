package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"imessage-exporter-webservice/utils"
)

// ProcessChatDB handles extracting messages from chat.db
func ProcessChatDB(w http.ResponseWriter, r *http.Request) {
	log.Println("📥 Process request received")

	// Parse JSON request body
	var request struct {
		S3URL        string `json:"s3_url"`
		PhoneNumber  string `json:"phone_number"`
		ContactName  string `json:"contact_name"`
		RequesterName string `json:"requester_name"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("❌ Error parsing request: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Define proper directory structure: /var/imessage/downloads/<RequesterName>/chat.db
	tempDir := "/var/imessage"
	localChatDBDir := filepath.Join(tempDir, "downloads", request.RequesterName)

	// Ensure the requester-specific directory exists
	err = os.MkdirAll(localChatDBDir, 0777)
	if err != nil {
		log.Printf("❌ ERROR: Failed to create directory %s - %v", localChatDBDir, err)
		http.Error(w, "Failed to create directory for processing", http.StatusInternalServerError)
		return
	}

	// Download chat.db from S3
	log.Printf("⬇️ Downloading chat.db from S3 for requester: %s\n", request.RequesterName)
	localChatDB, err := utils.DownloadFromS3(request.S3URL, request.RequesterName)
	if err != nil {
		log.Printf("❌ Error downloading chat.db: %v\n", err)
		http.Error(w, "Failed to download chat.db", http.StatusInternalServerError)
		return
	}

	// Create output directory specific to the requester
	outputDir := filepath.Join(tempDir, "output", request.RequesterName)
	err = os.MkdirAll(outputDir, 0777)
	if err != nil {
		log.Printf("❌ ERROR: Failed to create output directory %s - %v", outputDir, err)
		http.Error(w, "Failed to create output directory", http.StatusInternalServerError)
		return
	}

	// Run imessage-exporter for both html and txt
	err = runImessageExporter(localChatDB, outputDir, request.PhoneNumber)
	if err != nil {
		log.Printf("❌ Error processing chat.db: %v\n", err)
		http.Error(w, "Failed to process chat.db", http.StatusInternalServerError)
		return
	}

	// Upload extracted files to S3
	files, err := utils.UploadFolderToS3(outputDir, request.RequesterName)
	if err != nil {
		log.Printf("❌ Error uploading extracted files to S3: %v\n", err)
		http.Error(w, "Failed to upload extracted messages", http.StatusInternalServerError)
		return
	}

	// Return response
	response := map[string]interface{}{
		"message": "Processing complete",
		"files":   files,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func runImessageExporter(localChatDB, outputDir, phoneNumber string) error {
	formats := []string{"html", "txt"}

	for _, format := range formats {
		cmd := exec.Command("imessage-exporter",
			"-p", localChatDB,
			"-t", phoneNumber,
			"-o", outputDir,
			"-f", format)

		// ✅ Capture both stdout & stderr
		output, err := cmd.CombinedOutput()
		log.Printf("⚡ Running imessage-exporter with format: %s...\nCommand: imessage-exporter -p %s -t %s -o %s -f %s\nOutput: %s",
			format, localChatDB, phoneNumber, outputDir, format, string(output))

		if err != nil {
			log.Printf("❌ imessage-exporter error (format: %s): %v\nOutput: %s", format, err, string(output))
			return err
		}
	}
	return nil
}

