package converters

import (
	"context"
	"io"
)

type Converter interface {
	Convert(ctx context.Context, reader io.ReadCloser) (string, error)
}
