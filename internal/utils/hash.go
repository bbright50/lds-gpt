package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func GetSHA256Hash(reader io.ReadSeeker) (string, error) {
	// if reader is nil, return a default hash (verified by "sha256sum" command)
	if reader == nil {
		return "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", nil
	}

	h := sha256.New()

	// read the data from the reader into the hash
	if _, err := io.Copy(h, reader); err != nil {
		return "", err
	}

	// reset the reader to the beginning
	reader.Seek(0, io.SeekStart)

	// get the hash sum as a byte slice
	hashInBytes := h.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}
