package server

import "os/exec"

// InstallPackages installs the packages.
func InstallPackages(packages []string) (string, error) {
	cmd := exec.Command("sudo", "apt-get", "update")
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	command := []string{"apt-get", "install", "-y"}
	command = append(command, packages...)
	cmd = exec.Command("sudo", command...)
	err = cmd.Run()
	if err != nil {
		out, _ := cmd.CombinedOutput()
		return string(out), err
	}
	return "", nil
}
