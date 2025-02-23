package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"slices"

	"github.com/recally-io/go-markitdown"
	"github.com/recally-io/go-markitdown/converters"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

// Supported input formats and schemes
const (
	defaultModel = "gpt-4o-mini"
)

var (
	supportedExtensions = []string{".pdf", ".html", ".htm"}
	supportedSchemes    = []string{"http://", "https://"}
)

// Command line flags
type convertFlags struct {
	outputPath string
	model      string
}

func main() {
	// Initialize slog
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	if err := newRootCmd().Execute(); err != nil {
		slog.Error("command execution failed", "error", err)
		os.Exit(1)
	}
}

// newRootCmd creates the root command
func newRootCmd() *cobra.Command {
	flags := &convertFlags{}
	cmd := &cobra.Command{
		Use:   "markitdown <file|url>",
		Short: "Convert documents to markdown",
		Long:  `A powerful document conversion tool that preserves semantic structure`,
		Example: `  # Convert a local PDF file
  markitdown document.pdf -o output.md

  # Convert from a URL
  markitdown https://example.com/document.html -o output.md

  # Use a specific LLM model
  markitdown document.pdf -m gpt-4 -o output.md

  # Output to stdout (no -o flag)
  markitdown document.pdf`,
		Args: cobra.ExactArgs(1),
		RunE: flags.runConvert,
	}

	// Define flags
	cmd.Flags().StringVarP(&flags.outputPath, "output", "o", "", "Output file path")
	cmd.Flags().StringVarP(&flags.model, "model", "m", defaultModel, "LLM model to use")

	return cmd
}

// runConvert handles the conversion logic
func (f *convertFlags) runConvert(cmd *cobra.Command, args []string) error {
	input := args[0]

	if err := validateInput(input); err != nil {
		return err
	}

	// Show progress
	slog.Info("Converting file to markdown", "input", input)
	startTime := time.Now()

	// Convert document
	text, err := f.convertDocument(input)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Handle output
	if err := f.writeOutput(text); err != nil {
		return err
	}

	slog.Info("Conversion completed", "duration", time.Since(startTime).Round(time.Millisecond))
	return nil
}

// convertDocument performs the actual document conversion
func (f *convertFlags) convertDocument(input string) (string, error) {
	ctx := context.Background()
	converter := initializeConverter(f.model)
	return converter.Convert(ctx, input)
}

// initializeConverter creates and configures the converter
func initializeConverter(model string) *markitdown.MarkitDown {
	llmCfg := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	llmCfg.BaseURL = os.Getenv("OPENAI_BASE_URL")
	llmClient := openai.NewClientWithConfig(llmCfg)

	return markitdown.NewMarkitDown(
		converters.WithLLMClient(llmClient),
		converters.WithLLMModel(model),
	)
}

// validateInput checks if the input is valid
func validateInput(input string) error {
	if isURL(input) {
		return validateURL(input)
	}
	return validateFile(input)
}

// isURL checks if the input is a URL
func isURL(input string) bool {
	for _, scheme := range supportedSchemes {
		if strings.HasPrefix(input, scheme) {
			return true
		}
	}
	return false
}

// validateURL checks if the URL scheme is supported
func validateURL(url string) error {
	if !isURL(url) {
		return fmt.Errorf("invalid URL scheme: must be http or https")
	}
	return nil
}

// validateFile checks if the file exists and has a supported extension
func validateFile(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("input file not found: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	if slices.Contains(supportedExtensions, ext) {
		return nil
	}
	return fmt.Errorf("unsupported file type: %s (supported: %v)", ext, supportedExtensions)
}

// writeOutput writes the converted text to file or stdout
func (f *convertFlags) writeOutput(text string) error {
	if f.outputPath == "" {
		fmt.Println(text)
		return nil
	}

	if err := os.WriteFile(f.outputPath, []byte(text), 0600); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	return nil
}
