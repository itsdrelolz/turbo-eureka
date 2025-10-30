package gemini

import (
	"context"

	"google.golang.org/genai"
)

type ResumerParser interface {
	GetResumeText(ctx context.Context, resume []byte) (*genai.GenerateContentResponse, error)
}
