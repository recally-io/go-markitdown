package html

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/go-shiori/go-readability"
	"github.com/recally-io/go-markitdown/converters"
)

type Converter struct {
	options *converters.Options
}

func NewConverter(opts ...converters.Option) *Converter {
	return &Converter{
		options: converters.NewOptions(opts...),
	}
}

func (c *Converter) Convert(ctx context.Context, reader io.ReadCloser) (string, error) {
	defer reader.Close()

	var htmlContent []byte
	var err error

	if c.options.HtmlReadability {
		slog.Info("parsing HTML with readability mode enabled", "host", c.options.HtmlHost)
		article, err := readability.FromReader(reader, nil)
		if err != nil {
			slog.Error("failed to parse HTML content", "error", err)
			return "", fmt.Errorf("failed to parse HTML content: %w", err)
		}
		htmlContent = []byte(article.Content)
	} else {
		slog.Info("parsing raw HTML content", "host", c.options.HtmlHost)
		htmlContent, err = io.ReadAll(reader)
		if err != nil {
			slog.Error("failed to read HTML content", "error", err)
			return "", fmt.Errorf("failed to read HTML content: %w", err)
		}
	}

	converter := md.NewConverter(c.options.HtmlHost, true, &md.Options{})
	markdown, err := converter.ConvertString(string(htmlContent))
	if err != nil {
		slog.Error("markdown conversion failed", "error", err)
		return "", fmt.Errorf("webreader markdown converter error: %w", err)
	}

	slog.Info("completed HTML conversion", "content_length", len(markdown))
	return markdown, nil
}
