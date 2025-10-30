package gemini

import (
	"context"

	"google.golang.org/genai"
)

type ResumerParser interface {
	ExtractText(ctx context.Context, resume []byte) (*genai.GenerateContentResponse, error)
}
