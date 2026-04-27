//go:build !windows

package env

import "syscall"

// execProcess uses the Unix exec syscall to replace the current process.
func execProcess(path string, argv []string, env []string) error {
	return syscall.Exec(path, argv, env)
}
