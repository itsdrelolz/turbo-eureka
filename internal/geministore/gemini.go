package geministore

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/genai"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GeminiClient struct {
	Client *genai.Client
}

func New(ctx context.Context, apiKey string) (*GeminiClient, error) {

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})

	if err != nil {
		return nil, fmt.Errorf("API key error: %w", err)
	}

	return &GeminiClient{Client: client}, nil
}

func (g *GeminiClient) ExtractText(ctx context.Context, resume []byte) (string, error) {

	contents := []*genai.Content{
		genai.NewContentFromBytes(resume, "application/pdf", genai.RoleUser),
	}

	var invalidKey *genai.APIError

	result, err := g.Client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		contents,
		nil,
	)

	if err != nil {
		st, ok := status.FromError(err)

		if ok {

			switch st.Code() {

			case codes.Unauthenticated, codes.PermissionDenied:
				return "", fmt.Errorf("gemini authentication failed: %w", ErrPermanentFailure)

			case codes.InvalidArgument:
				return "", fmt.Errorf("gemini invalud input (400): %w", ErrPermanentFailure)
			}
		}

		return "", fmt.Errorf("Failed to extract text from resume with: %w", err)
	}

	return result.Text(), nil

}

func (g *GeminiClient) Embed(ctx context.Context, resumeText string) ([]float32, error) {

	contents := []*genai.Content{
		genai.NewContentFromText(resumeText, genai.RoleUser),
	}

	result, err := g.Client.Models.EmbedContent(ctx,
		"gemini-embedding-001",
		contents,
		nil,
	)
	if err != nil {

		st, ok := status.FromError(err)
		if ok {
			// 🛑 Permanent Error Check 2: Bad credentials or bad input
			switch st.Code() {
			case codes.Unauthenticated, codes.PermissionDenied: // Invalid API key or scope
				return nil, fmt.Errorf("gemini authentication failed: %w", ErrPermanentFailure)
			case codes.InvalidArgument: // Bad input (e.g., file too big, invalid format)
				return nil, fmt.Errorf("gemini invalid input (400): %w", ErrPermanentFailure)
			}
		}
		return nil, fmt.Errorf("Failed to embed given content with: %w", err)
	}

	if len(result.Embeddings) == 0 || len(result.Embeddings[0].Values) == 0 {
		return nil, fmt.Errorf("gemini returned empty embedding result")
	}

	vector := result.Embeddings[0].Values

	return vector, nil
}
