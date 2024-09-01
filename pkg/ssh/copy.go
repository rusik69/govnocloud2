package ssh

import (
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Copy copies files from source to destination.
func Copy(src, dst, host, user, key string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	keyBody, err := os.ReadFile(key)
	if err != nil {
		return err
	}
	singer, err := ssh.ParsePrivateKey(keyBody)
	if err != nil {
		return err
	}
	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(singer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", host+":22", &config)
	if err != nil {
		return err
	}
	defer client.Close()
	sftp, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	dstFile, err := sftp.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		return err
	}
	return nil
}
