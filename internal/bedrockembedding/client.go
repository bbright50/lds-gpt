package bedrockembedding

import (
	"context"
	"encoding/json"
	"fmt"
	"lds-gpt/internal/utils/rate_limiter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

const EMBEDDING_MODEL_ID = "amazon.titan-embed-text-v2:0"

//go:generate mockgen -source=client.go -destination=mocks/mock_bedrock_embedding_client.go -package=mocks
type Client interface {
	EmbedText(ctx context.Context, text string) ([]float64, error)
}

type clientConfig struct {
	maxConcurrentRequests int
}

type clientConfigOption func(*clientConfig)

func WithMaxConcurrentRequests(maxConcurrentRequests int) clientConfigOption {
	return func(opts *clientConfig) {
		opts.maxConcurrentRequests = maxConcurrentRequests
	}
}

type embedRequest struct {
	InputText string `json:"inputText"`
}

type embedResponse struct {
	Embedding []float64 `json:"embedding"`
}

type client struct {
	*rate_limiter.Embeddable[[]float64]

	embedder *bedrockruntime.Client
}

func NewClient(awsConfig aws.Config, options ...clientConfigOption) Client {
	bedrockClient := bedrockruntime.NewFromConfig(awsConfig)

	return &client{
		Embeddable: rate_limiter.NewEmbeddable[[]float64](20),

		embedder: bedrockClient,
	}
}

func (a *client) EmbedText(ctx context.Context, text string) ([]float64, error) {
	return a.SubmitErr(func() ([]float64, error) {
		return a.embedText(ctx, text)
	})
}

func (a *client) embedText(ctx context.Context, text string) ([]float64, error) {
	body, err := json.Marshal(embedRequest{InputText: text})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embed request: %w", err)
	}

	resp, err := a.embedder.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(EMBEDDING_MODEL_ID),
		Body:        body,
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return nil, err
	}

	var result embedResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embed response: %w", err)
	}
	if result.Embedding == nil {
		return nil, fmt.Errorf("embedding not found in response")
	}
	return result.Embedding, nil
}
