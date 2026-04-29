// Package template resolves secret references embedded in string templates.
// References use the syntax {{ vault:secret/data/myapp#key }} or
// {{ ssm:/myapp/key }}.
package template

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Resolver fetches a secret value given a provider prefix and path.
type Resolver interface {
	GetSecret(ctx context.Context, provider, path string) (string, error)
}

// refPattern matches {{ provider:path#key }} or {{ provider:path }}
var refPattern = regexp.MustCompile(`\{\{\s*(\w+):([^}#\s]+)(?:#([^}\s]+))?\s*\}\}`)

// Expand replaces all secret references in s with their resolved values.
// Unknown providers or missing secrets cause an error.
func Expand(ctx context.Context, s string, r Resolver) (string, error) {
	var firstErr error
	result := refPattern.ReplaceAllStringFunc(s, func(match string) string {
		if firstErr != nil {
			return match
		}
		parts := refPattern.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		provider := parts[1]
		path := parts[2]
		if len(parts) == 4 && parts[3] != "" {
			path = path + "#" + parts[3]
		}
		val, err := r.GetSecret(ctx, provider, path)
		if err != nil {
			firstErr = fmt.Errorf("template: resolving %q: %w", match, err)
			return match
		}
		return val
	})
	if firstErr != nil {
		return "", firstErr
	}
	return result, nil
}

// HasRefs reports whether s contains any secret references.
func HasRefs(s string) bool {
	return refPattern.MatchString(s)
}

// ListRefs returns all unique raw references found in s.
func ListRefs(s string) []string {
	matches := refPattern.FindAllString(s, -1)
	seen := make(map[string]struct{}, len(matches))
	out := matches[:0]
	for _, m := range matches {
		key := strings.TrimSpace(m)
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			out = append(out, key)
		}
	}
	return out
}
