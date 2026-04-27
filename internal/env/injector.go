package env

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Injector merges resolved secrets into the process environment
// and can exec a child process with that environment.
type Injector struct {
	baseEnv []string
}

// NewInjector creates an Injector seeded with the current process environment.
func NewInjector() *Injector {
	return &Injector{baseEnv: os.Environ()}
}

// Merge returns a copy of the base environment with the provided secrets
// overlaid. Keys in secrets override existing environment variables.
func (inj *Injector) Merge(secrets map[string]string) []string {
	result := make([]string, 0, len(inj.baseEnv)+len(secrets))

	// Track which keys secrets will override so we can skip them.
	override := make(map[string]struct{}, len(secrets))
	for k := range secrets {
		override[strings.ToUpper(k)] = struct{}{}
	}

	for _, entry := range inj.baseEnv {
		parts := strings.SplitN(entry, "=", 2)
		if _, ok := override[strings.ToUpper(parts[0])]; !ok {
			result = append(result, entry)
		}
	}

	for k, v := range secrets {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// Exec replaces the current process image with the given command, injecting
// the provided secrets into its environment. On Unix this uses exec(2);
// on other platforms it falls back to running a child process and waiting.
func (inj *Injector) Exec(secrets map[string]string, argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("env: argv must not be empty")
	}

	env := inj.Merge(secrets)

	path, err := exec.LookPath(argv[0])
	if err != nil {
		return fmt.Errorf("env: command not found: %w", err)
	}

	return execProcess(path, argv, env)
}
