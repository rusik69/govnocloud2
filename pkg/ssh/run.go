package ssh

import (
	"bufio"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

// Run runs the command on the remote host and streams the output.
func Run(cmd, host, key, user string, stream bool) (string, error) {
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
	var res string

	if stream {
		// Get the stdout and stderr pipes
		stdout, err := session.StdoutPipe()
		if err != nil {
			return "", err
		}
		stderr, err := session.StderrPipe()
		if err != nil {
			return "", err
		}

		// Start the command
		if err := session.Start(cmd); err != nil {
			return "", err
		}

		// Stream stdout
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				log.Println(scanner.Text())
			}
		}()

		// Stream stderr
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				log.Println(scanner.Text())
			}
		}()

		// Wait for the command to finish
		if err := session.Wait(); err != nil {
			return "", err
		}
	} else {
		// Run the command
		output, err := session.CombinedOutput(cmd)
		if err != nil {
			return "", err
		}
		res = string(output)
	}
	return res, nil
}
