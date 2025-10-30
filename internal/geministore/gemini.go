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

func (g *GeminiClient) GetResumeText(ctx context.Context, resume []byte) (*genai.GenerateContentResponse, error) {

	parts := []*genai.Part{
		&genai.Part{
			InlineData: &genai.Blob{
				MIMEType: "application/pdf",
				Data:     resume,
			},
		},
		genai.NewPartFromText("Summarize this document"),
	}
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	result, err := g.Client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		contents,
		nil,
	)
	if err != nil {

		return nil, fmt.Errorf("Failed to extract text from resume with: %w", err)
	}

	return result, nil

}
