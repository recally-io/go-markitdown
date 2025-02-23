package markitdown

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/recally-io/go-markitdown/converters"
)

type MarkitDown struct {
	options []converters.Option
}

func NewMarkitDown(opts ...converters.Option) *MarkitDown {
	return &MarkitDown{
		options: opts,
	}
}

func (m *MarkitDown) Convert(ctx context.Context, source string) (string, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "file://") {
		return m.ConvertURL(ctx, source)
	}
	return m.ConvertLocal(ctx, source)
}

func (m *MarkitDown) ConvertLocal(ctx context.Context, filePath string) (string, error) {
	fileType, err := getFileTypeFromPath(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to determine file type: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	converter, err := NewConverter(fileType, m.options...)
	if err != nil {
		return "", fmt.Errorf("failed to create converter for type %s: %w", fileType, err)
	}

	markdown, err := converter.Convert(ctx, file)
	if err != nil {
		return "", fmt.Errorf("failed to convert %s to Markdown: %w", fileType, err)
	}

	return markdown, nil
}

func (m *MarkitDown) ConvertURL(ctx context.Context, uri string) (string, error) {

	u, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	fileType, err := getFileType(resp, u.String())
	if err != nil {
		return "", fmt.Errorf("failed to determine file type: %w", err)
	}

	options := append(m.options, converters.WithHtmlHost(u.Host))
	converter, err := NewConverter(fileType, options...)
	if err != nil {
		return "", fmt.Errorf("failed to create converter for type %s: %w", fileType, err)
	}

	markdown, err := converter.Convert(ctx, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to convert %s to Markdown: %w", fileType, err)
	}

	return markdown, nil
}
