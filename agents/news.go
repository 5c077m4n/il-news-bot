package agents

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/5c077m4n/il-news-bot/agents/feeds"
	"github.com/openai/openai-go"
)

func lefty(ctx context.Context, prompt string) (*AnchorResponse, error) {
	ynetFeed, err := feeds.GetYNet(ctx)
	if err != nil {
		return nil, err
	}

	response, err := llmQuery[AnchorResponse](
		ctx,
		openai.SystemMessage(`
			Act as a progressive news anchor who is principled, calm,
			and meticulous.
			Your perspective leans left—prioritizing social justice,
			environmental protection,
			and economic equality—but your primary allegiance is to the truth.
			Every headline you deliver must be accompanied by a specific,
			credible source.
			Avoid hyperbole; let the data and the ethics of the story drive the
			narrative. Your tone is professional, empathetic, and intellectually
			rigorous.
		`),
		openai.SystemMessage(
			"**Do not** send a headline without at least one link to the original source (the more souces the better).",
		),
		openai.SystemMessage(fmt.Sprintf("YNet articles: %s", ynetFeed)),
		openai.UserMessage(prompt),
	)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "fetched left news successfully", slog.Any("response", response))

	return response, nil
}

func righty(ctx context.Context, prompt string) (*AnchorResponse, error) {
	israelHayomFeed, err := feeds.GetIsrealHayom(ctx)
	if err != nil {
		return nil, err
	}

	response, err := llmQuery[AnchorResponse](
		ctx,
		openai.SystemMessage(`
			Act as a principled, center-right news anchor.
			Your tone is professional, traditional, and analytical.
			You prioritize individual liberty, fiscal responsibility, and local
			governance. Crucially, every headline must be followed by a specific,
			credible source or data point. Avoid hyperbole; focus on interpreting
			current events through a conservative lens while maintaining strict
			journalistic integrity and factual accuracy.
		`),
		openai.SystemMessage(
			"**Do not** send a headline without at least one link to the original source (the more souces the better).",
		),
		openai.SystemMessage(fmt.Sprintf("Israel Hayom articles: %s", israelHayomFeed)),
		openai.UserMessage(prompt),
	)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "fetched right news successfully", slog.Any("response", response))

	return response, nil
}

func accumilator(
	ctx context.Context,
	leftReponse, rightResoponse *AnchorResponse,
) (*AnchorResponse, error) {
	slog.InfoContext(
		ctx,
		"accumilating news",
		slog.Any("left", leftReponse),
		slog.Any("righty", rightResoponse),
	)

	anchorMessage := openai.SystemMessage(`
	# You are a fact checker:
	- Make sure that any and all information passed through you is true
	- Make sure that all links are valid and return a non-error status code (2**) when opening, that stories are mentioned more than once (a good indication but not definitive)
	- Use ONLY the links provided without adding new ones on your own

	# How to respond
	After validating all the news lists then you'll return only one that
	includes all good items from both of them without duplications (if a
	story is in more than one article then just attach all relevant links).
	In case you recieve a nil/empty list of news make sure to mention it in
	your response.
	Try to group the news results so most responses will have more than one link
	with an appropriet title and description.
	`)
	aritclesPrompt := openai.UserMessage(fmt.Sprintf(`
	<left_news_articles>%s</left_news_articles>
	<right_news_articles>%s</right_news_articles>
	`, leftReponse, rightResoponse))

	response, err := llmQuery[AnchorResponse](ctx, anchorMessage, aritclesPrompt)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "accumilated news successfully", slog.Any("response", response))

	return response, nil
}
