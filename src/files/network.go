package files

// files/network.go is for handling file uploads/downloads.

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func SendFileToClient(w http.ResponseWriter, filePath string) error {
	// Open the file to be sent
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Set the appropriate headers
	fileName := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")

	// Stream the file to the response
	_, err = io.Copy(w, file)
	return err
}
