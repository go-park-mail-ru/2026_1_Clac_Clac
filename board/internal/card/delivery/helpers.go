package delivery

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

const cardBytesContainContentType = 512

func GetCardContentType(file *os.File) (string, error) {
	buffer := make([]byte, cardBytesContainContentType)

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
