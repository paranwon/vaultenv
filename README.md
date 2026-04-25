# vaultenv

> Inject secrets from HashiCorp Vault or AWS SSM into process environments — without writing to disk.

---

## Installation

```bash
go install github.com/yourorg/vaultenv@latest
```

Or download a pre-built binary from the [releases page](https://github.com/yourorg/vaultenv/releases).

---

## Usage

Prefix any command with `vaultenv` to have secrets injected as environment variables at runtime.

**HashiCorp Vault:**
```bash
vaultenv --source vault --path secret/data/myapp -- ./myapp serve
```

**AWS SSM Parameter Store:**
```bash
vaultenv --source ssm --path /myapp/prod -- ./myapp serve
```

Secrets are fetched at process startup, injected into the child process environment, and never written to disk or shell history.

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--source` | Secret backend (`vault` or `ssm`) | `vault` |
| `--path` | Secret path or prefix to fetch | *(required)* |
| `--addr` | Vault server address | `$VAULT_ADDR` |
| `--region` | AWS region (SSM only) | `$AWS_REGION` |
| `--prefix` | Strip path prefix from env var names | `false` |

### Authentication

- **Vault:** Uses standard Vault environment variables (`VAULT_ADDR`, `VAULT_TOKEN`, etc.)
- **SSM:** Uses the default AWS credential chain (env vars, `~/.aws/credentials`, IAM role)

---

## How It Works

`vaultenv` fetches secrets from the specified backend, maps each key to an uppercase environment variable, then `exec`s your command as a child process with those variables set. The parent process exits immediately — secrets exist only in the child's memory.

---

## License

MIT © yourorg