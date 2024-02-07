package files

import (
	"os"
	"path/filepath"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Returns true if the path exists and is NOT a directory
func FileExists(file string) bool {
	info, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Returns true if all files exist, false otherwise
func FilesExist(files []string) bool {
	for _, file := range files {
		if !FileExists(file) {
			return false
		}
	}
	return true
}

// Returns true if the path exists and is a directory
func DirExists(dir string) bool {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// Creates a directory if it doesn't exist
func CreateDirIfNotExists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// Returns a slice of all files in the directory
func ListFiles(dir string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
