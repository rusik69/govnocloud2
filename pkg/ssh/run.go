package ssh

import (
	"os"

	"golang.org/x/crypto/ssh"
)

// Run runs the command on the remote host.
func Run(cmd, host, key, user string) (string, error) {
	// Read the private key file
	keyBody, err := os.ReadFile(key)
	if err != nil {
		return "", err
	}
	// Parse the private key
	signer, err := ssh.ParsePrivateKey(keyBody)
	if err != nil {
		return "", err
	}
	// Configure the SSH client
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// Connect to the remote host
	client, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		return "", err
	}
	defer client.Close()
	// Run the command
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}
	return string(output), err
}
