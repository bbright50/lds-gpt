package utils

import (
	"fmt"
	"io"
	"os"
)

func CreateTempFileFromReader(filePrefix string, reader io.Reader) (*os.File, error) {
	tmpDir := os.TempDir()
	tmpFilePattern := fmt.Sprintf("%s-*", filePrefix)
	f, err := os.CreateTemp(tmpDir, tmpFilePattern)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(f, reader)
	if err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}

	// Seek back to the beginning so the file is ready to read
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}

	return f, nil
}
