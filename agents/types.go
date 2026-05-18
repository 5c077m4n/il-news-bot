package agents

import (
	"fmt"
	"strings"
)

type sanitizationResponse struct {
	Prompt string `json:"prompt" jsonschema:"The prompt after sanitisation (in case that the prompt is dirty leave this empty)"`
	Reason string `json:"reason" jsonschema:"This is the reason for allowing/blocking the original prompt"`
	Score  uint8  `json:"score"  jsonschema:"Safety score (out of 10),minimum=0,maximum=10"`
}

type NewsItem struct {
	Title       string   `json:"title"       jsonschema:"The article's title"`
	Description string   `json:"description" jsonschema:"The article's description"`
	Links       []string `json:"links"       jsonschema:"The article's list of origin links"`
}

func (n *NewsItem) String() string {
	var linksBuilder strings.Builder
	for i, link := range n.Links {
		if i > 0 {
			linksBuilder.WriteString(" | ")
		}
		fmt.Fprintf(&linksBuilder, `<a href="%s">%d</a>`, link, i+1)
	}
	return fmt.Sprintf(
		"<strong>%s</strong>\n%s\n%s\n",
		n.Title,
		n.Description,
		linksBuilder.String(),
	)
}

type AnchorResponse struct {
	List []*NewsItem `json:"list" jsonschema:"Article list"`
}

func (a *AnchorResponse) String() string {
	articleBuilder := strings.Builder{}
	for _, item := range a.List {
		articleBuilder.WriteString(item.String() + "\n")
	}
	return articleBuilder.String()
}
