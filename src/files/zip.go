package files

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ZipDir(source string, dest string) error {
	// Verify that the source directory exists
	if !DirExists(source) {
		return errors.New("source directory does not exist or is not a directory")
	}

	// Create the destination zip file
	zipfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	// Create a zip writer
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	baseDir := filepath.Base(source)

	// Walk through the source directory and add all files to the zip
	err = filepath.WalkDir(source, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = baseDir + "/" + strings.TrimPrefix(path, source)

		if d.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !d.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			_, err = io.Copy(writer, file)
			file.Close()
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func UnZipDir(source, dest string) error {
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(dest, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		fileReader, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		if _, err = io.Copy(outFile, fileReader); err != nil {
			outFile.Close()
			fileReader.Close()
			return err
		}

		outFile.Close()
		fileReader.Close()
	}

	return nil
}

func ZipFile(source, dest string) error {
	// Verify that the source file exists
	if !FileExists(source) {
		return errors.New("source file does not exist")
	}

	// Create the destination zip file
	zipfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	// Create a zip writer
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	// Open the source file
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new file in the zip archive
	writer, err := archive.Create(filepath.Base(source))
	if err != nil {
		return err
	}

	// Copy the source file to the zip archive
	_, err = io.Copy(writer, file)
	if err != nil {
		return err
	}

	return nil
}

func UnZipFile(source, dest string) error {
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	if len(reader.File) != 1 {
		return errors.New("zip file contains more than one file")
	}

	file := reader.File[0]
	if file.FileInfo().IsDir() {
		return errors.New("zip file contains a directory")
	}

	// Create the destination file
	outFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Open the file in the zip archive
	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer fileReader.Close()

	// Copy the file to the destination
	_, err = io.Copy(outFile, fileReader)
	if err != nil {
		return err
	}

	return nil
}
