package agents

import (
	"context"
	"time"

	"github.com/mmcdole/gofeed"
)

type YNetFetchTool struct{}

func (t YNetFetchTool) Name() string {
	return "ynet_fetch_tool"
}
func (t YNetFetchTool) Description() string {
	return `
	Fetch slightly left-leaning news outlet YNet.
	Political orientation: -3 (On a scale of [-10, 10])
	`
}
func (t YNetFetchTool) Call(ctx context.Context, _input string) (string, error) {
	parserCtx, parserCancel := context.WithTimeout(ctx, 10*time.Second)
	defer parserCancel()

	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(
		"https://www.ynet.co.il/Integration/StoryRss2.xml",
		parserCtx,
	)
	if err != nil {
		return "", err
	}

	return feed.String(), nil
}

type IsraelHayomFetchTool struct{}

func (t IsraelHayomFetchTool) Name() string {
	return "israel_hayom_fetch_tool"
}
func (t IsraelHayomFetchTool) Description() string {
	return `
	Fetch slightly right leaning news outlet Israel Hayom.
	Political orientation: 3 (On a scale of [-10, 10])
	`
}
func (t IsraelHayomFetchTool) Call(ctx context.Context, _input string) (string, error) {
	parserCtx, parserCancel := context.WithTimeout(ctx, 10*time.Second)
	defer parserCancel()

	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(
		"https://www.israelhayom.co.il/rss.xml",
		parserCtx,
	)
	if err != nil {
		return "", err
	}

	return feed.String(), nil
}
