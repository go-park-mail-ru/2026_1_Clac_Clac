package delivery

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

const bytesContainContentType = 512

func GetContentType(file *os.File) (string, error) {
	buffer := make([]byte, bytesContainContentType)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("file.Read ContentType: %w", err)
	}

	contentType := http.DetectContentType(buffer[:n])

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("file.Seek: %w", err)
	}

	return contentType, nil
}
