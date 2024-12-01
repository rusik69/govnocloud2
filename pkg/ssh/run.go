package ssh

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// RunConfig holds configuration for SSH command execution
type RunConfig struct {
	Command  string
	Host     string
	KeyPath  string
	User     string
	Password string
	Stream   bool
	Timeout  time.Duration
	Port     int
}

// NewRunConfig creates a new run configuration with defaults
func NewRunConfig(cmd, host, key, user, password string, stream bool, timeout int) *RunConfig {
	return &RunConfig{
		Command:  cmd,
		Host:     host,
		KeyPath:  key,
		User:     user,
		Password: password,
		Stream:   stream,
		Timeout:  time.Duration(timeout) * time.Second,
		Port:     22,
	}
}

// Run executes a command on a remote host
func Run(cmd, host, key, user, password string, stream bool, timeout int) (string, error) {
	cfg := NewRunConfig(cmd, host, key, user, password, stream, timeout)
	return cfg.Execute()
}

// Execute performs the command execution
func (r *RunConfig) Execute() (string, error) {
	// Create SSH client
	client, err := r.createSSHClient()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer client.Close()

	// Create session
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	if r.Stream {
		return r.executeStreaming(session)
	}
	return r.executeNonStreaming(session)
}

// createSSHClient creates a new SSH client
func (r *RunConfig) createSSHClient() (*ssh.Client, error) {
	config, err := r.createClientConfig()
	if err != nil {
		return nil, err
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", r.Host, r.Port), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to host: %w", err)
	}

	return client, nil
}

// createClientConfig creates SSH client configuration
func (r *RunConfig) createClientConfig() (*ssh.ClientConfig, error) {
	var auth []ssh.AuthMethod

	if r.Password != "" {
		auth = append(auth, ssh.Password(r.Password))
	} else {
		signer, err := r.loadPrivateKey()
		if err != nil {
			return nil, err
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	return &ssh.ClientConfig{
		User:            r.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         r.Timeout,
	}, nil
}

// loadPrivateKey loads and parses the private key
func (r *RunConfig) loadPrivateKey() (ssh.Signer, error) {
	keyBytes, err := os.ReadFile(r.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return signer, nil
}

// executeStreaming executes a command with output streaming
func (r *RunConfig) executeStreaming(session *ssh.Session) (string, error) {
	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := session.Start(r.Command); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Stream output in goroutines
	errChan := make(chan error, 2)
	go r.streamOutput(stdout, "stdout", errChan)
	go r.streamOutput(stderr, "stderr", errChan)

	// Wait for command completion and check for streaming errors
	if err := session.Wait(); err != nil {
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	// Check for streaming errors
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			return "", fmt.Errorf("output streaming error: %w", err)
		}
	}

	return "", nil
}

// executeNonStreaming executes a command and returns its output
func (r *RunConfig) executeNonStreaming(session *ssh.Session) (string, error) {
	output, err := session.CombinedOutput(r.Command)
	if err != nil {
		return string(output), fmt.Errorf("command execution failed: %w", err)
	}
	return string(output), nil
}

// streamOutput streams command output to the logger
func (r *RunConfig) streamOutput(pipe io.Reader, name string, errChan chan<- error) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		log.Printf("[%s] %s", name, scanner.Text())
	}
	errChan <- scanner.Err()
}
