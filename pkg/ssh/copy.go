package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// CopyConfig holds configuration for file copy operations
type CopyConfig struct {
	Source      string
	Destination string
	Host        string
	User        string
	KeyPath     string
	Port        int
}

// NewCopyConfig creates a new copy configuration with defaults
func NewCopyConfig(src, dst, host, user, key string) *CopyConfig {
	return &CopyConfig{
		Source:      src,
		Destination: dst,
		Host:        host,
		User:        user,
		KeyPath:     key,
		Port:        22,
	}
}

// Copy copies files from source to destination using SFTP
func Copy(src, dst, host, user, key string) error {
	cfg := NewCopyConfig(src, dst, host, user, key)
	return cfg.Execute()
}

// Execute performs the file copy operation
func (c *CopyConfig) Execute() error {
	// Open source file
	srcFile, err := os.Open(c.Source)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create SSH client
	sshClient, err := c.createSSHClient()
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer sshClient.Close()

	// Create SFTP client
	sftpClient, err := c.createSFTPClient(sshClient)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Create destination directory if it doesn't exist
	if err := c.ensureDestinationDir(sftpClient); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create destination file
	dstFile, err := sftpClient.Create(c.Destination)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy file contents
	if err := c.copyContents(srcFile, dstFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}

// createSSHClient creates a new SSH client
func (c *CopyConfig) createSSHClient() (*ssh.Client, error) {
	// Read private key
	keyBytes, err := os.ReadFile(c.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	// Parse private key
	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Create SSH client config
	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to remote host
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to remote host: %w", err)
	}

	return client, nil
}

// createSFTPClient creates a new SFTP client
func (c *CopyConfig) createSFTPClient(sshClient *ssh.Client) (*sftp.Client, error) {
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}
	return sftpClient, nil
}

// ensureDestinationDir ensures the destination directory exists
func (c *CopyConfig) ensureDestinationDir(sftpClient *sftp.Client) error {
	destDir := filepath.Dir(c.Destination)
	if destDir == "." {
		return nil
	}

	// Create all parent directories
	if err := sftpClient.MkdirAll(destDir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	return nil
}

// copyContents copies data from source to destination
func (c *CopyConfig) copyContents(src io.Reader, dst io.Writer) error {
	_, err := io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}
	return nil
}
