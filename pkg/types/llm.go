package types

type LLM struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	Image     string `json:"image"`
	Memory    int    `json:"memory"`
	CPU       int    `json:"cpu"`
}

var LLMTypes = map[string]LLM{
	"llama3": LLM{
		Name:   "llama3",
		Image:  "llama3",
		Memory: 1024,
		CPU:    1,
	},
}
