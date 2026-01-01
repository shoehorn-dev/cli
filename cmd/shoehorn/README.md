# Shoehorn CLI

Command-line interface for the Shoehorn Internal Developer Portal.

## Installation

### Build from source

```bash
cd cmd/cli
go build -o shoehorn.exe .
```

### Add to PATH (optional)

```bash
# Linux/macOS
export PATH=$PATH:/path/to/shoehorn/bin

# Windows PowerShell
$env:Path += ";C:\Users\check\projects\shoehorn\cmd\cli\shoehorn.exe"
```

## Quick Start

### 1. Login

```bash
# Device flow (default, works in SSH/headless environments)
shoehorn auth login --device

# Specify custom server and issuer
shoehorn auth login --server https://alpha.shoehorn.dev 
```

The CLI will display:
```
🔐 Logging in to Shoehorn...

To authenticate, visit:
  http://localhost:9090/realms/shoehorn/device

And enter code: ABCD-1234

Code expires in 900 seconds

⏳ Waiting for authentication...
```

Go to the URL, enter the code, and authenticate. Once complete:

```
✓ Logged in as: user@example.com
✓ Authentication successful!
✓ Profile: default
✓ Server: http://localhost:8080
```

### 2. Check authentication status

```bash
shoehorn auth status
```

Output:
```
Profile: default
Server:  http://localhost:8080
Status:  ✓ Authenticated
Email:   user@example.com
Name:    John Doe
Tenant:  acme-corp
Token:   Valid (expires in 12 minutes)
```

### 3. List workflow runs

```bash
# Table format (default)
shoehorn forge run list

# JSON format
shoehorn forge run list --output json
```

### 4. Get run details

```bash
shoehorn forge run get <run-id>

# JSON format
shoehorn forge run get <run-id> --output json
```

### 5. Logout

```bash
shoehorn auth logout
```

## Commands

### Authentication

```bash
# Login with device flow
shoehorn auth login --device

# Check status
shoehorn auth status

# Logout
shoehorn auth logout
```

### Forge

```bash
# List runs
shoehorn forge run list

# Get run details
shoehorn forge run get <run-id>
```

## Configuration

Config file location: `~/.shoehorn/config.yaml`

Example config:
```yaml
version: "1.0"
current_profile: default

profiles:
  default:
    name: Production
    server: https://api.shoehorn.dev
    auth:
      issuer: https://auth.shoehorn.dev/realms/shoehorn
      client_id: shoehorn-cli
      access_token: eyJ...
      refresh_token: def...
      token_type: Bearer
      expires_at: "2025-10-22T10:45:00Z"
      user:
        email: user@example.com
        name: John Doe
        tenant_id: acme-corp

  staging:
    name: Staging
    server: https://api.staging.shoehorn.dev
    auth:
      # ... auth details ...
```

## Multiple Profiles

```bash
# Login to different environments
shoehorn auth login --profile prod --server https://api.shoehorn.dev
shoehorn auth login --profile staging --server https://api.staging.shoehorn.dev
shoehorn auth login --profile dev --server http://localhost:8080

# Use specific profile
shoehorn --profile staging forge run list
```

## Environment Variables

```bash
# Override server URL
export SHOEHORN_SERVER=https://api.custom.dev

# Override profile
export SHOEHORN_PROFILE=staging
```

## Development

### Run from source

```bash
cd cmd/cli
go run . auth status
```

### Build

```bash
# Windows
go build -o ../../bin/shoehorn.exe .

# Linux/macOS
go build -o ../../bin/shoehorn .
```

### Add new commands

1. Create command file in `cmd/cli/commands/`
2. Add command to parent command in `init()` function
3. Implement `RunE` function
4. Rebuild CLI

Example:
```go
var myCmd = &cobra.Command{
    Use:   "my-command",
    Short: "My command description",
    RunE:  runMyCommand,
}

func init() {
    rootCmd.AddCommand(myCmd)
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

## Features

### Implemented

- ✅ OAuth2 device flow authentication
- ✅ Config file management (`~/.shoehorn/config.yaml`)
- ✅ Multiple authentication profiles
- ✅ Auth status checking
- ✅ Logout
- ✅ Forge run list
- ✅ Forge run get (details)
- ✅ Table and JSON output formats
- ✅ Token-based API authentication

### Coming Soon

- ⏳ Browser-based authentication flow (PKCE)
- ⏳ Automatic token refresh
- ⏳ Forge run create
- ⏳ Forge run logs (streaming)
- ⏳ Mold management commands
- ⏳ Interactive mode
- ⏳ Shell completion

## Troubleshooting

### "Not authenticated" error

```bash
# Check current status
shoehorn auth status

# Re-authenticate
shoehorn auth login --device
```

### Token expired

The CLI currently doesn't auto-refresh tokens. Re-authenticate:

```bash
shoehorn auth login --device
```

### API connection errors

Check server URL in config:
```bash
cat ~/.shoehorn/config.yaml
```

Update server URL:
```bash
shoehorn auth login --device --server http://localhost:8080
```

### Device code expired

Device codes typically expire after 15 minutes. If you see "expired_token" error, start a new login:

```bash
shoehorn auth login --device
```

## Architecture

```
cmd/cli/
├── main.go              # Entry point
├── commands/
│   ├── root.go          # Root command + global flags
│   ├── auth.go          # Auth commands (login, status, logout)
│   └── forge.go         # Forge commands (run list, run get)
├── config/
│   └── config.go        # Config file management
└── api/
    └── client.go        # HTTP API client
```

### Authentication Flow

1. User runs `shoehorn auth login --device`
2. CLI calls Keycloak device authorization endpoint
3. Keycloak returns device code + user code
4. CLI displays verification URL and user code
5. User visits URL in browser and enters code
6. CLI polls token endpoint every 5 seconds
7. Once authorized, CLI receives access token + refresh token
8. Tokens stored in `~/.shoehorn/config.yaml`
9. Future API calls use access token in Authorization header

### API Client

Simple HTTP client with:
- Automatic Bearer token authentication
- GET and POST methods
- JSON request/response handling
- Error handling with status codes

## Contributing

1. Create feature branch
2. Add command/feature
3. Test with `go run . <command>`
4. Build and test binary
5. Update README
6. Submit PR

## License

[Your License]
