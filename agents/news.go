package agents

import (
	"context"
	"log/slog"

	agentTools "github.com/5c077m4n/il-news-bot/agents/tools"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

func lefty(ctx context.Context, prompt string) ([]*NewsItem, error) {
	leftPromptTemplate := prompts.PromptTemplate{Template: prompt}

	leftyTools := []tools.Tool{agentTools.YNetFetchTool{}}
	response, err := llmQuery[AnchorResponse](ctx, leftPromptTemplate, leftyTools)
	if err != nil {
		return nil, err
	}

	return response.List, nil
}

func righty(ctx context.Context, prompt string) ([]*NewsItem, error) {
	rightyPromptTemplate := prompts.PromptTemplate{Template: prompt}

	rightyTools := []tools.Tool{agentTools.IsraelHayomFetchTool{}}
	response, err := llmQuery[AnchorResponse](ctx, rightyPromptTemplate, rightyTools)
	if err != nil {
		return nil, err
	}

	return response.List, nil
}

func accumilator(
	ctx context.Context,
	leftReponse, rightResoponse []*NewsItem,
) ([]*NewsItem, error) {
	slog.InfoContext(ctx, "accumilating news", "left", leftReponse, "righty", rightResoponse)

	accuPropmptTemplate := prompts.PromptTemplate{
		Template: `
		<left_news_articles>{{ .lefty }}</left_news_articles>
		<right_news_articles>{{ .righty }}</right_news_articles>
		`,
		InputVariables:   []string{"lefty", "righty"},
		TemplateFormat:   prompts.TemplateFormatGoTemplate,
		PartialVariables: map[string]any{"lefty": leftReponse, "righty": rightResoponse},
	}

	accuTools := []tools.Tool{}
	response, err := llmQuery[AnchorResponse](ctx, accuPropmptTemplate, accuTools)
	if err != nil {
		return nil, err
	}
	return response.List, nil
}
