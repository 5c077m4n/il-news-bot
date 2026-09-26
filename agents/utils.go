package agents

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	openrouter "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/OpenRouterTeam/go-sdk/optionalnullable"
	"github.com/goccy/go-json"
	"github.com/google/jsonschema-go/jsonschema"
)

const deepseekModel = "deepseek/deepseek-v4-flash"
const jevModel = "typesafe/jev-1.13"

var ErrLLMReponseParse = errors.New("could not prase the LLM's resposne")
var client = openrouter.New(
	openrouter.WithSecurity(os.Getenv("OPENROUTER_API_KEY")),
)

func systemMessage(content string) components.ChatMessages {
	return components.CreateChatMessagesSystem(components.ChatSystemMessage{
		Role: components.ChatSystemMessageRoleSystem,
		Content: components.ChatSystemMessageContent{
			Str:  &content,
			Type: components.ChatSystemMessageContentTypeStr,
		},
	})
}

func userMessage(content string) components.ChatMessages {
	return components.CreateChatMessagesUser(components.ChatUserMessage{
		Role: components.ChatUserMessageRoleUser,
		Content: components.ChatUserMessageContent{
			Str:  &content,
			Type: components.ChatUserMessageContentTypeStr,
		},
	})
}

func llmQuery[T any](
	ctx context.Context,
	prompt components.ChatMessages,
	rest ...components.ChatMessages,
) (*T, error) {
	schema, err := jsonschema.For[T](nil)
	if err != nil {
		return nil, err
	}
	schemaBytes, err := json.MarshalContext(ctx, schema)
	if err != nil {
		return nil, err
	}

	messages := []components.ChatMessages{
		systemMessage(
			fmt.Sprintf(
				`Make sure to ALWAYS return your response in this JSON schema exactly and exclusively:
				<output_shape>%s</output_shape>`,
				string(schemaBytes),
			),
		),
		systemMessage(fmt.Sprintf("The current time is: %s", time.Now().String())),
		prompt,
	}
	if len(rest) > 0 {
		messages = append(messages, rest...)
	}
	params := components.ChatRequest{
		Messages: messages,
		Seed:     optionalnullable.From(new(int64(0))),
		Model:    new(deepseekModel),
	}

	for i := range 3 {
		completion, err := client.Chat.Send(ctx, params, nil)
		if err != nil {
			slog.WarnContext(
				ctx,
				"could not retrieve the LLM's response",
				slog.String("error", err.Error()),
				slog.Int("attempt", i),
				slog.String("model", deepseekModel),
			)
			time.Sleep(750 * time.Millisecond)
			continue
		}
		if completion == nil || completion.ChatResult == nil ||
			len(completion.ChatResult.Choices) == 0 {
			slog.WarnContext(
				ctx,
				"empty response from the LLM",
				slog.Int("attempt", i),
				slog.String("model", deepseekModel),
			)
			time.Sleep(750 * time.Millisecond)
			continue
		}

		var content string
		choice := completion.ChatResult.Choices[0].Message
		if c, ok := choice.Content.Get(); ok && c.Str != nil {
			content = *c.Str
		}

		var result T
		if err := json.UnmarshalContext(
			ctx,
			[]byte(content),
			&result,
		); err != nil {
			slog.WarnContext(
				ctx,
				"could not parse the LLM's response",
				slog.String("error", err.Error()),
				slog.Int("attempt", i),
				slog.String("model", deepseekModel),
			)
			time.Sleep(750 * time.Millisecond)
			continue
		}
		return &result, nil
	}

	return nil, ErrLLMReponseParse
}
