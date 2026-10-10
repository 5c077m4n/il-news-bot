package agents

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/5c077m4n/il-news-bot/db"
	"github.com/OpenRouterTeam/go-sdk/models/components"
)

const articlesPerAnchor = 20

func articlesMessage(articles []db.Article) components.ChatMessages {
	var contentBuilder strings.Builder
	contentBuilder.WriteString("Articles from our news sources:")
	for _, article := range articles {
		fmt.Fprintf(
			&contentBuilder,
			"\n- %s\n%s\n%s (%s)",
			article.Title,
			article.Description,
			article.Link,
			article.Source,
		)
	}
	return systemMessage(contentBuilder.String())
}

func anchorArticles(
	ctx context.Context,
	database *db.Database,
	prompt string,
) ([]components.ChatMessages, error) {
	articles, err := database.QueryArticles(ctx, prompt, articlesPerAnchor, nil)
	if err != nil {
		return nil, err
	}

	messages := []components.ChatMessages{}
	if len(articles) > 0 {
		messages = append(messages, articlesMessage(articles))
	}
	return messages, nil
}

func anchor(ctx context.Context, database *db.Database, prompt string) (*AnchorResponse, error) {
	messages, err := anchorArticles(ctx, database, prompt)
	if err != nil {
		return nil, err
	}
	messages = append(messages, userMessage(prompt))
	response, err := llmQuery[AnchorResponse](
		ctx,
		systemMessage(`
			Act as a principled news anchor who is calm and meticulous.
			Your tone is professional, neutral, and analytical.
			Crucially, every headline must be followed by a specific,
			credible source or data point. Avoid hyperbole; let the data
			drive the narrative while maintaining strict journalistic
			integrity and factual accuracy.
			**Do not** send a headline without at least one link to the original source (the more sources the better).
			Name every link with up to 2 words that describe its content.
		`),
		messages...,
	)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "fetched news successfully", slog.Any("response", response))

	return response, nil
}

func factChecker(
	ctx context.Context,
	language string,
	response *AnchorResponse,
) (*AnchorResponse, error) {
	slog.InfoContext(ctx, "checking news", slog.Any("response", response))

	anchorMessage := systemMessage(`
	# You are a fact checker:
	- Make sure that any and all information passed through you is true
	- Make sure that all links are valid and return a non-error status code (2**) when opening, that stories are mentioned more than once (a good indication but not definitive)
	- Use ONLY the links provided without adding new ones on your own
	- Name every link with up to 2 words that describe its content

	# How to respond
	After validating the news list then you'll return only one that includes all good items without duplications (if a
	story is in more than one article then just attach all relevant links). In case you recieve a nil/empty list of news make sure to mention it in your response.
	Try to group the news results by subject so most responses will have more than one link with an appropriet title and up to 20 word description.
	Also translate the response to the requested language (if given).
	Return at most 5 news groups.
	`)
	articlesPrompt := userMessage(
		fmt.Sprintf(
			`<news_articles>%s</news_articles><requested_language>%s</requested_language>`,
			response,
			language,
		),
	)

	checked, err := llmQuery[AnchorResponse](ctx, anchorMessage, articlesPrompt)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "checked news successfully", slog.Any("response", checked))

	return checked, nil
}
