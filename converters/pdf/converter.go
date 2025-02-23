package pdf

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/gen2brain/go-fitz"
	"github.com/recally-io/go-markitdown/converters"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/sync/errgroup"
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
	doc, err := fitz.NewFromReader(reader)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	var texts []string
	if c.options.LLMClient != nil {
		slog.Info("converting PDF with LLM assistance")
		texts, err = c.ConvertPagesWithLLM(ctx, doc)
	} else {
		slog.Info("converting PDF without LLM")
		texts, err = c.ConvertPages(ctx, doc)
	}
	if err != nil {
		slog.Error("failed to convert pages", "error", err)
		return "", fmt.Errorf("failed to convert pages: %w", err)
	}

	return strings.Join(texts, "\n\n"), nil
}

func (c *Converter) ConvertPages(ctx context.Context, doc *fitz.Document) ([]string, error) {
	totalPages := doc.NumPage()
	slog.Info("starting PDF conversion", "pages", totalPages)

	results := make([]string, totalPages)
	for i := range results {
		slog.Debug("extracting text from page", "page", i+1, "total", totalPages)
		pageText, err := doc.Text(i)
		if err != nil {
			slog.Error("failed to extract text", "page", i+1, "error", err)
			return nil, fmt.Errorf("failed to extract text from page %d: %w", i+1, err)
		}
		results[i] = pageText
	}

	slog.Info("completed PDF conversion", "pages", totalPages)
	return results, nil
}

func (c *Converter) ConvertPagesWithLLM(ctx context.Context, doc *fitz.Document) ([]string, error) {
	totalPages := doc.NumPage()
	slog.Info("starting PDF conversion with LLM", "pages", totalPages, "workers", c.options.NumWorkers)

	g, ctx := errgroup.WithContext(ctx)
	results := make([]string, totalPages)
	var mu sync.Mutex

	// Process pages in parallel with bounded concurrency
	sem := make(chan struct{}, c.options.NumWorkers)
	for i := range results {
		i := i // Create new variable for goroutine

		g.Go(func() error {
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			slog.Debug("processing page with LLM", "page", i+1, "total", totalPages)
			markdown, err := c.processPage(ctx, doc, i)
			if err != nil {
				slog.Error("failed to process page", "page", i+1, "error", err)
				return fmt.Errorf("failed to process page %d: %w", i, err)
			}

			mu.Lock()
			results[i] = markdown
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		slog.Error("page processing failed", "error", err)
		return nil, fmt.Errorf("page processing failed: %w", err)
	}

	slog.Info("completed PDF conversion with LLM", "pages", totalPages)
	return results, nil
}

// processPage handles the conversion of a single PDF page
func (c *Converter) processPage(ctx context.Context, doc *fitz.Document, pageNum int) (string, error) {
	// Extract text
	slog.Debug("extracting text from page", "page", pageNum)
	pageText, err := doc.Text(pageNum)
	if err != nil {
		slog.Error("failed to extract text", "page", pageNum, "error", err)
		return "", fmt.Errorf("failed to extract text from page %d: %w", pageNum, err)
	}

	// Extract image
	slog.Debug("extracting image from page", "page", pageNum)
	img, err := doc.ImagePNG(pageNum, c.options.ImageDPI)
	if err != nil {
		slog.Error("failed to extract image", "page", pageNum, "error", err)
		return "", fmt.Errorf("failed to extract image from page %d: %w", pageNum, err)
	}

	// Create chat completion request
	req := openai.ChatCompletionRequest{
		Model: c.options.LLMModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: c.options.LLMPrompt,
			},
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: fmt.Sprintf("<pageContent>%s</pageContent>", pageText),
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: fmt.Sprintf("data:image/png;base64,%s",
								base64.StdEncoding.EncodeToString(img)),
						},
					},
				},
			},
		},
	}

	// Get response from API
	slog.Debug("calling LLM API", "page", pageNum)
	resp, err := c.options.LLMClient.CreateChatCompletion(ctx, req)
	if err != nil {
		slog.Error("API call failed", "page", pageNum, "error", err)
		return "", fmt.Errorf("API call failed for page %d: %w", pageNum, err)
	}

	content := resp.Choices[0].Message.Content
	content = strings.TrimPrefix(content, "```markdown\n")
	content = strings.TrimSuffix(content, "\n```")

	return content, nil
}
