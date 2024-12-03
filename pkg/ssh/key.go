package ssh

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

// KeyConfig holds configuration for SSH key operations
type KeyConfig struct {
	Host       string
	User       string
	Password   string
	PublicKey  string
	MasterHost string
	KeyPath    string
	Port       int
	Timeout    int
}

// NewKeyConfig creates a new key configuration with defaults
func NewKeyConfig(host, user, password, pubKeyPath, master string) *KeyConfig {
	return &KeyConfig{
		Host:       host,
		User:       user,
		Password:   password,
		PublicKey:  pubKeyPath,
		MasterHost: master,
		Port:       22,
		Timeout:    10,
	}
}

// CopySSHKey copies the SSH key to the remote server
func CopySSHKey(host, user, password, pubKeyPath, master string) error {
	cfg := NewKeyConfig(host, user, password, pubKeyPath, master)
	return cfg.CopyKey()
}

// CopyKey performs the key copy operation
func (k *KeyConfig) CopyKey() error {
	if k.MasterHost == "" {
		return k.copyKeyDirect()
	}
	return k.copyKeyViaMaster()
}

// copyKeyDirect copies the key directly to the target host
func (k *KeyConfig) copyKeyDirect() error {
	// Read the public key file
	publicKey, err := os.ReadFile(k.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %w", err)
	}

	// Create SSH client
	client, err := k.createSSHClient()
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer client.Close()

	// Setup authorized keys
	if err := k.setupAuthorizedKeys(client, string(publicKey)); err != nil {
		return err
	}

	// Copy to root user
	if err := k.copyToRoot(client); err != nil {
		return err
	}

	return nil
}

// createSSHClient creates a new SSH client using password authentication
func (k *KeyConfig) createSSHClient() (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: k.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(k.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", k.Host, k.Port), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return client, nil
}

// setupAuthorizedKeys sets up the authorized_keys file
func (k *KeyConfig) setupAuthorizedKeys(client *ssh.Client, publicKey string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf(
		"mkdir -p ~/.ssh && echo %q >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys",
		publicKey,
	)

	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to setup authorized_keys: %w", err)
	}

	return nil
}

// copyToRoot copies the authorized_keys file to root user
func (k *KeyConfig) copyToRoot(client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	cmd := "sudo mkdir -p /root/.ssh && sudo cp ~/.ssh/authorized_keys /root/.ssh/authorized_keys"
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to copy keys to root: %w", err)
	}

	return nil
}

// copyKeyViaMaster copies the key through a master host
func (k *KeyConfig) copyKeyViaMaster() error {
	cmd := fmt.Sprintf("sshpass -p %s ssh-copy-id %s@%s", k.Password, k.User, k.Host)
	log.Printf("Running: %s", cmd)

	out, err := Run(cmd, k.MasterHost, k.PublicKey, k.User, k.Password, false, k.Timeout)
	if err != nil {
		return fmt.Errorf("failed to copy SSH key via master: %s: %w", out, err)
	}

	return nil
}

// CreateKey creates an SSH key pair if it doesn't exist
func CreateKey(host, path, user, key string) (string, error) {
	cfg := &KeyConfig{
		Host:    host,
		KeyPath: path,
		User:    user,
		Timeout: 10,
	}
	return cfg.CreateKeyPair()
}

// CreateKeyPair creates a new SSH key pair
func (k *KeyConfig) CreateKeyPair() (string, error) {
	cmd := fmt.Sprintf(
		`if [ ! -f %s ]; then ssh-keygen -t rsa -b 4096 -C %q -f %s -N ""; fi`,
		k.KeyPath,
		k.User,
		k.KeyPath,
	)
	log.Printf("Running: %s", cmd)

	out, err := Run(cmd, k.Host, k.KeyPath, k.User, "", false, k.Timeout)
	if err != nil {
		return string(out), fmt.Errorf("failed to create SSH key pair: %w", err)
	}

	return "", nil
}
