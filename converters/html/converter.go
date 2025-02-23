package html

import (
	"context"
	"fmt"
	"io"

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
		article, err := readability.FromReader(reader, nil)
		if err != nil {
			return "", fmt.Errorf("failed to parse HTML content: %w", err)
		}
		htmlContent = []byte(article.Content)
	} else {
		htmlContent, err = io.ReadAll(reader)
		if err != nil {
			return "", fmt.Errorf("failed to read HTML content: %w", err)
		}
	}

	converter := md.NewConverter(c.options.HtmlHost, true, &md.Options{})
	markdown, err := converter.ConvertString(string(htmlContent))
	if err != nil {
		return "", fmt.Errorf("webreader markdown converter error: %w", err)
	}
	return markdown, nil
}
