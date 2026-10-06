package db

import (
	"context"
	"os"
	"sync"

	"github.com/philippgille/chromem-go"
)

const openRouterBaseURL = "https://openrouter.ai/api/v1"
const embeddingModel = "openai/text-embedding-3-small"

var directory = "chromem_data"

var embeddingFunc chromem.EmbeddingFunc = func(ctx context.Context, text string) ([]float32, error) {
	return chromem.NewEmbeddingFuncOpenAICompat(
		openRouterBaseURL,
		os.Getenv("OPENROUTER_API_KEY"),
		embeddingModel,
		new(true),
	)(ctx, text)
}

type Database struct {
	articles *chromem.Collection
}

var database = sync.OnceValues(func() (*Database, error) {
	instance, err := chromem.NewPersistentDB(directory, false)
	if err != nil {
		return nil, err
	}

	articles, err := instance.GetOrCreateCollection("articles", nil, embeddingFunc)
	if err != nil {
		return nil, err
	}

	return &Database{articles: articles}, nil
})
