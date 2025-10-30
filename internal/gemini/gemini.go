package gemini

import (
	"context"
)

type ResumerParser interface {
	ExtractText(ctx context.Context, resume []byte) (string, error)
	Embed(ctz context.Context, resumeText string ) ([]byte, error)
}
