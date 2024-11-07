package ssh

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/crypto/ssh"
)

// CopySSHKey copies the SSH key to the remote server.
func CopySSHKey(host, user, password, pubKeyPath string) error {
	// Read the public key file
	publicKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return fmt.Errorf("unable to read public key file: %v", err)
	}

	// Create SSH client configuration
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password), // Use password authentication
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Insecure for demo purposes; use a proper callback in production
	}

	// Connect to the server
	conn, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Create a new session
	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// Prepare the command to append the public key to authorized_keys
	cmd := fmt.Sprintf(`mkdir -p ~/.ssh && echo "%s" >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys`, string(publicKey))

	// Run the command on the remote server
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to run command on remote server: %v", err)
	}

	// Copy the public key to the root user's authorized_keys file
	cmd = "sudo cp ~/.ssh/authorized_keys /root/.ssh/authorized_keys"
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to run command on remote server: %v", err)
	}

	return nil
}

// CreateKey creates the key.
func CreateKey(path string) (string, error) {
	// check if key exists
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			out, err := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", path, "-N", "").CombinedOutput()
			return string(out), err
		}
	} else {
		return "key already exists", nil
	}
	return "", nil
}
