package types

type LLM struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
}

var LLMTypes = map[string]LLM{
	"llama3": LLM{
		Name:      "llama3",
		Namespace: "llama3",
		Image:     "llama3",
		Memory:    1024,
		CPU:       1,
	},
}
