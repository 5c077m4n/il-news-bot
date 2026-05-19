package agents

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const model = "google/gemini-3.1-flash-lite-preview"

var client = openai.NewClient(
	option.WithBaseURL("https://openrouter.ai/api/v1"),
	option.WithAPIKey(os.Getenv("OPENROUTER_API_KEY")),
)

var ErrLLMReponseParse = errors.New("could not prase the LLM's resposne")

func llmQuery[T any](
	ctx context.Context,
	prompt openai.ChatCompletionMessageParamUnion,
	rest ...openai.ChatCompletionMessageParamUnion,
) (*T, error) {
	schema, err := jsonschema.For[T](nil)
	if err != nil {
		return nil, err
	}
	schemaBytes, err := json.MarshalContext(ctx, schema)
	if err != nil {
		return nil, err
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(fmt.Sprintf(`
		Make sure to ALWAYS return your response in this JSON schema exactly and exclusively:
		<output_shape>%s</output_shape>
		`, string(schemaBytes))),
		openai.SystemMessage(fmt.Sprintf("The current time is: %s", time.Now().String())),
		prompt,
	}
	if len(rest) > 0 {
		messages = append(messages, rest...)
	}
	params := openai.ChatCompletionNewParams{Messages: messages, Seed: openai.Int(0), Model: model}

	for i := range 3 {
		completion, err := client.Chat.Completions.New(ctx, params)
		if err != nil {
			slog.WarnContext(
				ctx,
				"could not retrieve the LLM's response",
				slog.String("error", err.Error()),
				slog.Int("attempt", i),
				slog.String("model", model),
			)
			time.Sleep(750 * time.Millisecond)
			continue
		}

		var result T
		if err := json.UnmarshalContext(
			ctx,
			[]byte(completion.Choices[0].Message.Content),
			&result,
		); err != nil {
			slog.WarnContext(
				ctx,
				"could not parse the LLM's response",
				slog.String("error", err.Error()),
				slog.Int("attempt", i),
				slog.String("model", model),
			)
			time.Sleep(750 * time.Millisecond)
			continue
		}
		return &result, nil
	}

	return nil, ErrLLMReponseParse
}
