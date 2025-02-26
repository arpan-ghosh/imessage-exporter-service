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
		RequesterName string `json:"requester_name"` // New field
	}
	

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("❌ Error parsing request: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create temporary directory
	tempDir := "/var/imessage"
	os.MkdirAll(tempDir, 0777)
	// Define proper directory structure: /var/imessage/downloads/<RequesterName>/chat.db
	localChatDBDir := filepath.Join(tempDir, "downloads", request.RequesterName)

	// Download chat.db from S3
	log.Printf("⬇️ Downloading chat.db from S3 into %s\n", localChatDBDir)
	err = utils.DownloadFromS3(request.S3URL, localChatDBDir, request.RequesterName)
	if err != nil {
		log.Printf("❌ Error downloading chat.db: %v\n", err)
		http.Error(w, "Failed to download chat.db", http.StatusInternalServerError)
		return
	}

	// Correct path where chat.db is stored
	localChatDB := filepath.Join(localChatDBDir, "chat.db")

	if err != nil {
		log.Printf("❌ Error downloading chat.db: %v\n", err)
		http.Error(w, "Failed to download chat.db", http.StatusInternalServerError)
		return
	}

	// Create output directory
	outputDir := filepath.Join(tempDir, "output")
	os.MkdirAll(outputDir, 0777)

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

// runImessageExporter runs imessage-exporter twice: once for HTML, once for TXT
func runImessageExporter(localChatDB, outputDir, phoneNumber string) error {
	formats := []string{"html", "txt"}

	for _, format := range formats {
		cmd := exec.Command("imessage-exporter",
			"-p", localChatDB,
			"-t", phoneNumber,
			"-o", outputDir,
			"-f", format)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		log.Printf("⚡ Running imessage-exporter with format: %s...", format)
		if err := cmd.Run(); err != nil {
			log.Printf("❌ Error running imessage-exporter (%s): %v", format, err)
			return err
		}
	}
	return nil
}
