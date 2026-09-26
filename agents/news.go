package agents

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/5c077m4n/il-news-bot/agents/feeds"
)

func lefty(ctx context.Context, prompt string) (*AnchorResponse, error) {
	var ynetFeed string

	var wg sync.WaitGroup
	wg.Go(func() {
		feed, err := feeds.GetYNet(ctx)
		if err != nil {
			slog.WarnContext(ctx, "could not fetch YNet data source", slog.Any("error", err))
		} else {
			ynetFeed = feed
		}
	})
	wg.Wait()

	response, err := llmQuery[AnchorResponse](
		ctx,
		systemMessage(`
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
		systemMessage(
			"**Do not** send a headline without at least one link to the original source (the more souces the better).",
		),
		systemMessage(fmt.Sprintf("YNet articles: %s", ynetFeed)),
		userMessage(prompt),
	)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "fetched left news successfully", slog.Any("response", response))

	return response, nil
}

func righty(ctx context.Context, prompt string) (*AnchorResponse, error) {
	var israelHayomFeed, jpostFeed, makorRishon string

	var wg sync.WaitGroup
	wg.Go(func() {
		feed, err := feeds.GetIsrealHayom(ctx)
		if err != nil {
			slog.WarnContext(
				ctx,
				"could not fetch data source",
				slog.String("source", "Israel Hayom"),
				slog.Any("error", err),
			)
		} else {
			israelHayomFeed = feed
		}
	})
	wg.Go(func() {
		feed, err := feeds.GetJPost(ctx)
		if err != nil {
			slog.WarnContext(
				ctx,
				"could not fetch data source",
				slog.String("source", "JPost"),
				slog.Any("error", err),
			)
		} else {
			jpostFeed = feed
		}
	})
	wg.Go(func() {
		feed, err := feeds.GetMakorRishon(ctx)
		if err != nil {
			slog.WarnContext(
				ctx,
				"could not fetch data source",
				slog.String("source", "Makor Rishon"),
				slog.Any("error", err),
			)
		} else {
			makorRishon = feed
		}
	})
	wg.Wait()

	response, err := llmQuery[AnchorResponse](
		ctx,
		systemMessage(`
			Act as a principled, center-right news anchor.
			Your tone is professional, traditional, and analytical.
			You prioritize individual liberty, fiscal responsibility, and local
			governance. Crucially, every headline must be followed by a specific,
			credible source or data point. Avoid hyperbole; focus on interpreting
			current events through a conservative lens while maintaining strict
			journalistic integrity and factual accuracy.
		`),
		systemMessage(
			"**Do not** send a headline without at least one link to the original source (the more souces the better).",
		),
		systemMessage(fmt.Sprintf("Israel Hayom articles: %s", israelHayomFeed)),
		systemMessage(fmt.Sprintf("Jerusalem Post articles: %s", jpostFeed)),
		systemMessage(fmt.Sprintf("Makor Rishon articles: %s", makorRishon)),
		userMessage(prompt),
	)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "fetched right news successfully", slog.Any("response", response))

	return response, nil
}

func accumilator(
	ctx context.Context,
	language string,
	leftReponse, rightResoponse *AnchorResponse,
) (*AnchorResponse, error) {
	slog.InfoContext(
		ctx,
		"accumilating news",
		slog.Any("left", leftReponse),
		slog.Any("righty", rightResoponse),
	)

	anchorMessage := systemMessage(`
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
	Try to group the news results so most responses will have more than one link with an appropriet title and description.
	Also translate the response to the requested language (if given).
	Return at most 5 news groups.
	`)
	aritclesPrompt := userMessage(
		fmt.Sprintf(
			`<left_news_articles>%s</left_news_articles><right_news_articles>%s</right_news_articles><requested_language>%s</requested_language>`,
			leftReponse,
			rightResoponse,
			language,
		),
	)

	response, err := llmQuery[AnchorResponse](ctx, anchorMessage, aritclesPrompt)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "accumilated news successfully", slog.Any("response", response))

	return response, nil
}
