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

func MakeRequest(ctx context.Context, userPrompt string, imagesB64 []string) (*openai_structs.ProductSEO, error) {
	apiKey := os.Getenv(defines.OpenAiApiKeyEnv)
	if apiKey == "" {
		return nil, fmt.Errorf("%s is not set", defines.OpenAiApiKeyEnv)
	}

	client := openaigit.NewClient(option.WithAPIKey(apiKey))

	systemMsg := openaigit.SystemMessage(DevPrompt)

	parts := make([]openaigit.ChatCompletionContentPartUnionParam, 0, 1+len(imagesB64))
	parts = append(parts, openaigit.ChatCompletionContentPartUnionParam{
		OfText: &openaigit.ChatCompletionContentPartTextParam{
			Text: userPrompt,
		},
	})

	for _, b64 := range imagesB64 {
		if strings.TrimSpace(b64) == "" {
			continue
		}
		imgURL := b64
		if !strings.HasPrefix(b64, "data:") {
			imgURL = "data:image/jpeg;base64," + b64
		}

		parts = append(parts, openaigit.ChatCompletionContentPartUnionParam{
			OfImageURL: &openaigit.ChatCompletionContentPartImageParam{
				ImageURL: openaigit.ChatCompletionContentPartImageImageURLParam{
					URL: imgURL,
				},
			},
		})
	}

	userMsg := openaigit.ChatCompletionMessageParamUnion{
		OfUser: &openaigit.ChatCompletionUserMessageParam{
			Content: openaigit.ChatCompletionUserMessageParamContentUnion{
				OfArrayOfContentParts: parts,
			},
		},
	}

	resp, err := client.Chat.Completions.New(ctx, openaigit.ChatCompletionNewParams{
		Model: openaigit.ChatModel(Model),
		Messages: []openaigit.ChatCompletionMessageParamUnion{
			systemMsg,
			userMsg,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("API error: %w", err)
	}

	if resp == nil || len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response received from model")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)

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
