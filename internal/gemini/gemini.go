package gemini

import (
	"context"
)

type ResumerParser interface {
	ExtractText(ctx context.Context, resume []byte) (string, error)
	Embed(ctx context.Context, resumeText string) ([]float32, error)
}
