package audit

// NoopLogger is a Logger that discards all events.
// Useful as a safe default when audit logging is not configured.
var NoopLogger = New(nil, false)

// Ensure *Logger satisfies the Auditor interface at compile time.
var _ Auditor = (*Logger)(nil)

// Auditor is the minimal interface consumers should depend on.
type Auditor interface {
	SecretFetched(provider, path, envVar string) error
	SecretError(provider, path, envVar string, fetchErr error) error
	ProcessExec(argv []string) error
	Log(e Event) error
}
