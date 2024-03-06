package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func extractTarGz(filePath, extractPath string) error {
	// Open the gzip file
	gzFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer gzFile.Close()

	// Create a gzip reader
	gzReader, err := gzip.NewReader(gzFile)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzReader)

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()

		// If no more files are found, break
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Construct the path to extract to
		targetPath := filepath.Join(extractPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create the directory with the same permissions as the tar header
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create the file with the same permissions as the tar header
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// Copy the file data from the tar archive to the file
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}
