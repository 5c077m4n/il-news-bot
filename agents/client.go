package agents

import (
	"errors"
	"os"

	openrouter "github.com/OpenRouterTeam/go-sdk"
)

const deepseekModel = "deepseek/deepseek-v4-flash"
const jevModel = "typesafe/jev-1.13"

var ErrLLMReponseParse = errors.New("could not prase the LLM's resposne")
var client = openrouter.New(
	openrouter.WithSecurity(os.Getenv("OPENROUTER_API_KEY")),
)
