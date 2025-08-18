package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Shu-AFK/WawiIC/cmd/defines"
	"github.com/Shu-AFK/WawiIC/cmd/openai/openai_structs"
	openaigit "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

func CheckForAPIKey() error {
	_, exists := os.LookupEnv(defines.OpenAiApiKeyEnv)
	if !exists {
		return fmt.Errorf("OpenAI API key not found in environment")
	}
	return nil
}

func MakeRequest(ctx context.Context, userPrompt string) (*openai_structs.ProductSEO, error) {
	apiKey := os.Getenv(defines.OpenAiApiKeyEnv)
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is not set")
	}

	client := openaigit.NewClient(option.WithAPIKey(apiKey))

	resp, err := client.Chat.Completions.New(ctx, openaigit.ChatCompletionNewParams{
		Model: openaigit.ChatModel(Model),
		Messages: []openaigit.ChatCompletionMessageParamUnion{
			openaigit.SystemMessage(DevPrompt),
			openaigit.UserMessage(userPrompt),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response received from model")
	}

	content := resp.Choices[0].Message.Content
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var out openai_structs.ProductSEO
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w\nRaw response:\n%s", err, content)
	}

	return &out, nil
}
