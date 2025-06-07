package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// CreateLLM creates a new LLM deployment
func (c *Client) CreateLLM(name, namespace, llmType string) error {
	llm := types.LLM{
		Name:      name,
		Namespace: namespace,
		Type:      llmType,
	}

	data, err := json.Marshal(llm)
	if err != nil {
		return fmt.Errorf("error marshaling LLM: %w", err)
	}

	url := fmt.Sprintf("%s/llms/%s/%s", c.baseURL, namespace, name)

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		return fmt.Errorf("error creating LLM: %s %s", resp.Status, string(body))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create LLM: %s %s", resp.Status, string(body))
	}

	return nil
}

// DeleteLLM deletes an LLM
func (c *Client) DeleteLLM(namespace, name string) error {
	url := fmt.Sprintf("%s/llms/%s/%s", c.baseURL, namespace, name)

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting LLM: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete LLM: %s %s", resp.Status, string(body))
	}
	return nil
}

// GetLLM gets an LLM
func (c *Client) GetLLM(name, namespace string) (types.LLM, error) {
	url := fmt.Sprintf("%s/llms/%s/%s", c.baseURL, namespace, name)

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return types.LLM{}, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return types.LLM{}, fmt.Errorf("error getting LLM: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.LLM{}, fmt.Errorf("error reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return types.LLM{}, fmt.Errorf("failed to get LLM: %s %s", resp.Status, string(body))
	}
	return types.LLM{}, nil
}

// ListLLMs lists all LLMs in a namespace
func (c *Client) ListLLMs(namespace string) ([]types.LLM, error) {
	url := fmt.Sprintf("%s/llms/%s", c.baseURL, namespace)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error listing LLMs: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list LLMs: %s %s", resp.Status, string(body))
	}

	llms := []types.LLM{}
	if err := json.Unmarshal(body, &llms); err != nil {
		return nil, fmt.Errorf("error unmarshalling LLMs: %w", err)
	}
	return llms, nil
}
