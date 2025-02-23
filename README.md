# go-markitdown
A CLI tool and library written in Go for converting documents to Markdown format.

## Features
- Convert PDF, HTML documents to Markdown
- Support for both local files and URLs
- Preserve semantic structure during conversion
- Easy to use CLI interface
- Flexible library integration

## Installation

```bash
go install github.com/recally-io/go-markitdown/cmd/markitdown@latest
```

## Usage

### CLI Usage

The `markitdown` command line tool provides a simple interface for document conversion:

#### Required Environment Variables

Before using the CLI tool, make sure to set the following environment variables:

```bash
export OPENAI_BASE_URL="https://api.openai.com/v1"  # Or your custom OpenAI API endpoint
export OPENAI_API_KEY="your-api-key-here"           # Your OpenAI API key
```

These environment variables are required for PDF text extraction using OpenAI's models.

```bash
# Convert a local file
markitdown convert document.pdf -o output.md

# Convert from URL
markitdown convert https://example.com/document.html -o output.md

# Preserve original document layout
markitdown convert document.pdf --preserve-layout -o output.md

# Specify a different LLM model
markitdown convert document.pdf -m gpt-4 -o output.md
```

Available flags:
- `-o, --output`: Output file path (if not specified, outputs to stdout)
- `-m, --model`: LLM model to use (default: gpt-4o-mini)
- `--preserve-layout`: Maintain original document layout

### Library Usage

To use go-markitdown as a library in your Go project:

```go
package main

import (
    "context"
    "github.com/recally-io/go-markitdown"
    "github.com/recally-io/go-markitdown/converters"
)

func main() {
    // Create a new MarkitDown instance with options
    md := markitdown.NewMarkitDown(
        converters.WithPreserveLayout(true),
        // Add other options as needed
    )

    // Convert a local file
    markdown, err := md.ConvertLocal(context.Background(), "document.pdf")
    if err != nil {
        // Handle error
    }

    // Convert from URL
    markdown, err = md.ConvertURL(context.Background(), "https://example.com/document.html")
    if err != nil {
        // Handle error
    }

    // Generic convert method (auto-detects source type)
    markdown, err = md.Convert(context.Background(), "document.pdf")
    if err != nil {
        // Handle error
    }
}
```

## Supported Formats

Input formats:
- PDF (.pdf)
- HTML (.html, .htm)
- URLs (http://, https://)

Output format:
- Markdown (.md)

## License

[MIT License](./LICENSE)
