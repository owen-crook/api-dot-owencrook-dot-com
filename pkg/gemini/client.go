package gemini

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type Client struct {
	apiKey string
	model  string
	client *genai.Client
}

func NewClient(ctx context.Context, apiKey string, model string) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Client{
		apiKey: apiKey,
		model:  model,
		client: client,
	}, nil
}

func (c *Client) GenerateFromTextAndImage(ctx context.Context, prompt string, imageBytes []byte) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("gemini client not initialized")
	}

	parts := []*genai.Part{
		genai.NewPartFromText(prompt),
		genai.NewPartFromBytes(imageBytes, "image/jpeg"),
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	result, err := c.client.Models.GenerateContent(ctx, c.model, contents, nil)
	if err != nil {
		return "", fmt.Errorf("gemini image + text generation failed: %w", err)
	}

	return result.Text(), nil
}
