package markitdown

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

// getFileType determines the file type from an HTTP response that can be used for conversion.
// It returns a simple file type string (e.g., "html", "pdf") that can be mapped to converters.
// Returns an error if the file type is unsupported or cannot be determined.
func getFileType(resp *http.Response, url string) (string, error) {
	if resp == nil {
		return "", fmt.Errorf("http response is nil")
	}

	// First try Content-Type header
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" {
		// Strip any charset or boundary information
		if idx := strings.Index(contentType, ";"); idx != -1 {
			contentType = contentType[:idx]
		}
		contentType = strings.TrimSpace(contentType)

		// Map MIME types to file types
		switch contentType {
		case "text/html", "application/xhtml+xml":
			return "html", nil
		case "application/pdf":
			return "pdf", nil
		case "application/epub+zip":
			return "epub", nil
		case "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
			return "doc", nil
		case "text/markdown":
			return "md", nil
		case "text/plain":
			return "txt", nil
		}
	}

	// Fallback to URL extension
	ext, err := getFileTypeFromPath(url)
	if err != nil {
		return "", fmt.Errorf("failed to determine file type from URL: %w", err)
	}
	return ext, nil
}

// getFileTypeFromPath determines the file type from a file path.
func getFileTypeFromPath(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".html", ".htm":
		return "html", nil
	case ".pdf":
		return "pdf", nil
	case ".epub":
		return "epub", nil
	case ".doc", ".docx":
		return "doc", nil
	case ".md", ".markdown":
		return "md", nil
	case ".txt":
		return "txt", nil
	}
	return "", fmt.Errorf("unsupported or unknown file type: %s", ext)
}
