package routes

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"tsm/src/files"
)

// Helper function to load and parse a template file.
func LoadTemplate(templatePath string) (*template.Template, error) {
	if !files.FileExists(templatePath) {
		return nil, fmt.Errorf("%s not found", filepath.Base(templatePath))
	}

	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}

	return template.New(filepath.Base(templatePath)).Parse(string(templateContent))
}
