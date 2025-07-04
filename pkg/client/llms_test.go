package client_test

import (
	"testing"
)

func TestCreateLLM(t *testing.T) {
	cli := setupTestClient(t)
	err := cli.CreateLLM("test-llm", testNamespace, "deepseek-r1-1.5b")
	if err != nil {
		t.Fatalf("error creating LLM: %v", err)
	}
}

func TestGetLLM(t *testing.T) {
	cli := setupTestClient(t)
	llm, err := cli.GetLLM("test-llm", testNamespace)
	if err != nil {
		t.Fatalf("error getting LLM: %v", err)
	}
	t.Logf("LLM: %v", llm)
}

func TestListLLMs(t *testing.T) {
	cli := setupTestClient(t)
	llms, err := cli.ListLLMs(testNamespace)
	if err != nil {
		t.Fatalf("error listing LLMs: %v", err)
	}
	t.Logf("LLMs: %v", llms)
}
