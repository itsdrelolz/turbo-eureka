package geministore

import (
	"context"
	"fmt"

	"google.golang.org/genai"
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

	result, err := g.Client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		contents,
		nil,
	)
	if err != nil {

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
		return nil, fmt.Errorf("Failed to embed given content with: %w", err)
	}

	if len(result.Embeddings) == 0 || len(result.Embeddings[0].Values) == 0 {
		return nil, fmt.Errorf("gemini returned empty embedding result")
	}

	vector := result.Embeddings[0].Values


	return vector, nil
}
