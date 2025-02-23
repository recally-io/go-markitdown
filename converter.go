package markitdown

import (
	"fmt"

	"github.com/recally-io/go-markitdown/converters"
	"github.com/recally-io/go-markitdown/converters/html"
	"github.com/recally-io/go-markitdown/converters/pdf"
)

func NewConverter(extension string, opts ...converters.Option) (converters.Converter, error) {
	switch extension {
	case "html":
		return html.NewConverter(opts...), nil
	case "pdf":
		return pdf.NewConverter(opts...), nil
	}
	return nil, fmt.Errorf("unsupported converter type: %s", extension)
}
