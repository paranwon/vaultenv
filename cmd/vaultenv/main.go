// Package main is the entry point for the vaultenv CLI tool.
// It wires together configuration loading, secret providers, caching,
// audit logging, and process execution.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yourorg/vaultenv/internal/audit"
	"github.com/yourorg/vaultenv/internal/cache"
	"github.com/yourorg/vaultenv/internal/config"
	"github.com/yourorg/vaultenv/internal/env"
	"github.com/yourorg/vaultenv/internal/masker"
	"github.com/yourorg/vaultenv/internal/provider"
)

const (
	// defaultCacheTTL is how long fetched secrets are cached in memory.
	defaultCacheTTL = 5 * time.Minute

	// exitCodeUsage is returned when the user provides invalid arguments.
	exitCodeUsage = 2
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "vaultenv: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		os.Exit(exitCodeUsage)
	}

	// Load configuration from the default search path or VAULTENV_CONFIG env var.
	cfg, err := config.LoadDefault()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Build the audit logger (writes JSON lines to stderr when enabled).
	auditor, err := audit.New(cfg.Audit)
	if err != nil {
		return fmt.Errorf("initialising audit logger: %w", err)
	}
	defer auditor.Close() //nolint:errcheck

	// Build the secret provider (Vault or SSM), wrapped in retry + cache layers.
	base, err := provider.New(cfg)
	if err != nil {
		return fmt.Errorf("creating provider: %w", err)
	}

	secretsCache := cache.New(defaultCacheTTL)
	cachedProv := cache.NewCachedProvider(base, secretsCache)

	// Resolve every mapping declared in the config file.
	ctx := context.Background()
	secrets := make(map[string]string, len(cfg.Mappings))

	for _, m := range cfg.Mappings {
		value, fetchErr := cachedProv.GetSecret(ctx, m.Path, m.Key)
		if fetchErr != nil {
			auditor.SecretError(m.EnvVar, fetchErr) //nolint:errcheck
			return fmt.Errorf("fetching secret for %s: %w", m.EnvVar, fetchErr)
		}
		auditor.SecretFetched(m.EnvVar, m.Path) //nolint:errcheck
		secrets[m.EnvVar] = value
	}

	// Build a masking writer so secret values never appear in plain text on
	// stderr should the child process print them.
	m := masker.New()
	for _, v := range secrets {
		m.Add(v)
	}
	maskedStderr := masker.NewWriter(os.Stderr, m)

	// Merge the resolved secrets into the current environment.
	injector := env.NewInjector(log.New(maskedStderr, "", 0))
	merged := injector.Merge(os.Environ(), secrets)

	// Record the exec event before handing over to the child process.
	auditor.ProcessExec(args) //nolint:errcheck

	// Replace the current process with the requested command.
	if err := injector.Exec(args, merged); err != nil {
		return fmt.Errorf("exec %q: %w", args[0], err)
	}

	// exec replaces the process; we only reach here on Windows where Exec
	// runs a child and waits.
	return nil
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: vaultenv <command> [args...]

vaultenv injects secrets from HashiCorp Vault or AWS SSM into the environment
of <command> without writing them to disk.

Configuration is read from vaultenv.yaml (or the file pointed to by
VAULTENV_CONFIG) in the current directory or $HOME/.config/vaultenv/.

Examples:
  vaultenv ./server
  vaultenv -- python app.py --port 8080`)
}
