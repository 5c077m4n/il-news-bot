package agents

import (
	"fmt"
	"strings"
)

type Link struct {
	URL  string `json:"url"  jsonschema:"The link's URL"`
	Name string `json:"name" jsonschema:"The link's name, up to 2 words describing the link's content"`
}

type NewsItem struct {
	Title       string `json:"title"       jsonschema:"The article's title"`
	Description string `json:"description" jsonschema:"The article's description"`
	Links       []Link `json:"links"       jsonschema:"The article's list of origin links"`
}

func (l *Link) String() string {
	words := strings.Fields(l.Name)
	if len(words) > 2 {
		words = words[:2]
	}
	name := strings.Join(words, " ")
	if name == "" {
		name = "link"
	}
	return fmt.Sprintf(`<a href="%s">%s</a>`, l.URL, name)
}

func (n *NewsItem) String() string {
	var linksBuilder strings.Builder
	for i, link := range n.Links {
		if i > 0 {
			linksBuilder.WriteString(" | ")
		}
		linksBuilder.WriteString(link.String())
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
