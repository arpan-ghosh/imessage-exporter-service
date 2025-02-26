package utils

import (
	"bytes"
	"fmt"
	"os/exec"
)

// RunExporter executes the imessage-exporter command
func RunExporter(dbPath, outputDir, phoneNumber, contactName string) error {
	cmd := exec.Command("imessage-exporter",
		"-f", "html",
		"-p", dbPath,
		"-t", phoneNumber,
		"-c", contactName,
		"--export-path", outputDir)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("❌ imessage-exporter error: %s\n", stderr.String())
		return err
	}

	fmt.Printf("✅ imessage-exporter output: %s\n", out.String())
	return nil
}
