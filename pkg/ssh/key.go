package ssh

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

// CopySSHKey copies the SSH key to the remote server.
func CopySSHKey(host, user, password, pubKeyPath, master string) error {
	if master == "" {
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
		cmd := fmt.Sprintf(`mkdir -p ~/.ssh && echo "%s" > ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys`, string(publicKey))

		// Run the command on the remote server
		if err := session.Run(cmd); err != nil {
			return fmt.Errorf("failed to run command on remote server: %v", err)
		}
		session, err = conn.NewSession()
		if err != nil {
			return fmt.Errorf("failed to create session: %v", err)
		}
		defer session.Close()
		// Copy the public key to the root user's authorized_keys file
		cmd = "sudo cp ~/.ssh/authorized_keys /root/.ssh/authorized_keys"
		if err := session.Run(cmd); err != nil {
			return fmt.Errorf("failed to run command on remote server: %v", err)
		}
	} else {
		cmd := fmt.Sprintf(`sshpass -p %s ssh-copy-id %s@%s`, password, user, host)
		log.Println(cmd)
		out, err := Run(cmd, master, pubKeyPath, user, password, false, 10)
		if err != nil {
			return fmt.Errorf("failed to copy ssh key: %v, %v", string(out), err)
		}
	}
	return nil
}

// CreateKey creates the key.
func CreateKey(host, path, user, key string) (string, error) {
	// check if key exists on the host and create if not
	cmd := fmt.Sprintf(`if [ ! -f %s ]; then ssh-keygen -t rsa -b 4096 -C "%s" -f %s -N ""; fi`, path, user, path)
	log.Println(cmd)
	out, err := Run(cmd, host, key, user, "", false, 10)
	if err != nil {
		return string(out), err
	}
	return "", nil
}
