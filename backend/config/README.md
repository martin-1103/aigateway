# Configuration

**Main config file:** `config/config.yaml`

This is the ONLY config file used by the backend. Do NOT create duplicate config files in other locations.

## Configuration Hierarchy:

1. **config/config.yaml** - Primary config (REQUIRED)
2. **Environment variables** - Override config values (optional)
3. **.env file** - Override config values (optional)

## Example:

```yaml
# config/config.yaml
auth_manager:
  enabled: true

# Can be overridden by:
# USE_AUTH_MANAGER=false (env var)
```

## See Also:

- `backend/.env.example` - Environment variable template
- `CLAUDE.md` - Full documentation
