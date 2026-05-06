package agents

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

var ErrLLMResposneParse = errors.New("could not parse LLM's response")

var getLLM = sync.OnceValues(func() (*openai.LLM, error) {
	return openai.New(
		openai.WithModel("google/gemini-3.1-flash-lite-preview"),
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
		openai.WithToken(os.Getenv("OPENROUTER_API_KEY")),
	)
})

func llmQuery[T any](
	ctx context.Context,
	partialPrompt prompts.PromptTemplate,
	toolList []tools.Tool,
) (*T, error) {
	schema, err := jsonschema.For[T](&jsonschema.ForOptions{IgnoreInvalidTypes: true})
	if err != nil {
		return nil, err
	}
	schemaJSON, err := json.MarshalContext(ctx, schema)
	if err != nil {
		return nil, err
	}

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	agent := agents.NewOneShotAgent(llm, toolList, agents.WithMaxIterations(3))
	executor := agents.NewExecutor(agent)

	partialPrompt.Template += `
	<current_time>{{ .timestamp }}</current_time>
	<output_shape_instructions>
		Make sure to return your results in this JSON schema form EXACTLY:
		<output_shape>{{ .output_shape }}</output_shape>
	</output_shape_instructions>
	`
	partialPrompt.InputVariables = append(partialPrompt.InputVariables, "timestamp", "output_shape")
	partialPrompt.TemplateFormat = prompts.TemplateFormatGoTemplate
	finalPrompt, err := partialPrompt.Format(
		map[string]any{"output_shape": string(schemaJSON), "timestamp": time.Now().String()},
	)
	if err != nil {
		return nil, err
	}

	for i := range 3 {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		llmResponse, err := chains.Run(ctx, executor, finalPrompt)
		if err != nil {
			slog.WarnContext(
				ctx,
				"could not run chain",
				slog.Int("iteration", i),
				slog.String("error", err.Error()),
			)
			continue
		}
		slog.InfoContext(
			ctx,
			"parsing response",
			slog.Int("attempt", i+1),
			slog.String("raw_reposne", llmResponse),
		)

		var structeredLLMResponse T
		if err := json.UnmarshalContext(
			ctx,
			[]byte(llmResponse),
			&structeredLLMResponse,
		); err != nil {
			slog.WarnContext(
				ctx,
				"LLM response parsing error",
				slog.String("error", err.Error()),
				slog.Int("attempt", i),
				slog.String("raw_reposne", llmResponse),
			)
			continue
		}

		return &structeredLLMResponse, nil
	}

	return nil, ErrLLMResposneParse
}
