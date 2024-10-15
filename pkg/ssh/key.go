package ssh

import (
	"os"
	"os/exec"
)

// InstallKey installs the key.
func InstallKey(host, user, password, key string) (string, error) {
	out, err := exec.Command("sshpass", "-p", password, "ssh-copy-id", "-i", key, user+"@"+host).CombinedOutput()
	return string(out), err
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
