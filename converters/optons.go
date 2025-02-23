package converters

import (
	_ "embed"

	"github.com/sashabaranov/go-openai"
)

//go:embed prompt.md
var systemMessage string

// Options represents the configuration options for the converter.
type Options struct {
	// LLMClient is the OpenAI client used for language model interactions.
	LLMClient *openai.Client

	// LLMPrompt is the system message or prompt used for the language model.
	LLMPrompt string

	// LLMModel specifies the language model to be used, default is "gpt-4o-mini".
	LLMModel string

	// NumWorkers determines the number of concurrent workers for processing.
	NumWorkers int

	// ImageDPI specifies the DPI for image extraction.
	ImageDPI float64

	// HtmlHost specifies the host for the HTML converter.
	HtmlHost string

	// HtmlReadability enables HTML readability mode to clean up the html before conversion.
	HtmlReadability bool
}

type Option func(*Options)

func NewOptions(opts ...Option) *Options {
	options := &Options{
		LLMPrompt:  systemMessage,
		LLMModel:   "gpt-4o-mini",
		NumWorkers: 10,
		ImageDPI:   300,

		HtmlReadability: true,
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

func WithLLMClient(client *openai.Client) Option {
	return func(o *Options) {
		o.LLMClient = client
	}
}

func WithLLMPrompt(prompt string) Option {
	return func(o *Options) {
		o.LLMPrompt = prompt
	}
}

func WithLLMModel(model string) Option {
	return func(o *Options) {
		o.LLMModel = model
	}
}

func WithNumWorkers(num int) Option {
	return func(o *Options) {
		o.NumWorkers = num
	}
}

func WithImageDPI(dpi float64) Option {
	return func(o *Options) {
		o.ImageDPI = dpi
	}
}

func WithHtmlHost(host string) Option {
	return func(o *Options) {
		o.HtmlHost = host
	}
}
