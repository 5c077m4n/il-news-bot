package db

import (
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/philippgille/chromem-go"
)

type Article struct {
	Source      string    `json:"source"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	PublishedAt time.Time `json:"publishedAt"`
}

func (a Article) content() string {
	return strings.TrimSpace(a.Title + "\n" + a.Description)
}

func (a Article) id() string {
	if a.Link != "" {
		return a.Link
	}

	hash := fnv.New64a()
	for _, part := range []string{a.Source, a.Title, a.Description} {
		_, _ = fmt.Fprintln(hash, part)
	}
	return fmt.Sprintf("%s:%x", a.Source, hash.Sum64())
}

func (a Article) metadata() map[string]string {
	return map[string]string{
		"source":      a.Source,
		"title":       a.Title,
		"description": a.Description,
		"link":        a.Link,
		"publishedAt": a.PublishedAt.Format(time.RFC3339),
	}
}

func from(result chromem.Result) Article {
	publishedAt, err := time.Parse(time.RFC3339, result.Metadata["publishedAt"])
	if err != nil {
		publishedAt = time.Time{}
	}

	return Article{
		Source:      result.Metadata["source"],
		Title:       result.Metadata["title"],
		Description: result.Metadata["description"],
		Link:        result.Metadata["link"],
		PublishedAt: publishedAt,
	}
}
