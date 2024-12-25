package ssh

import (
	"fmt"
	"log"
	"os/exec"
)

// Rsync copies files from source to destination using rsync
func Rsync(src, dst, host, user, key string) error {
	cmd := fmt.Sprintf("rsync --rsync-path='mkdir -p /var/www/govnocloud2 && rsync' -e 'ssh -i %s' -avz %s %s@%s:%s", key, src, user, host, dst)
	log.Println(cmd)
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to rsync: %s: %w", string(out), err)
	}
	log.Println(string(out))
	return nil
}
