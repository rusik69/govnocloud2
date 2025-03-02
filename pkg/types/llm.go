package types

type LLM struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
}

var LLMTypes = map[string]LLM{
	"deepseek-r1-1.5b": LLM{
		Type: "deepseek-r1-1.5b",
	},
	"deepseek-r1-7b": LLM{
		Type: "deepseek-r1-7b",
	},
}
