//go:build windows

package env

import (
	"os"
	"os/exec"
)

// execProcess spawns a child process on Windows (no exec(2) equivalent).
func execProcess(path string, argv []string, env []string) error {
	cmd := exec.Command(path, argv[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
