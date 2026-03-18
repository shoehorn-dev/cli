# Shoehorn CLI

Command-line interface for the Shoehorn Internal Developer Portal. Browse your service catalog, explore teams and ownership, run Forge workflows, and manage authentication — all from the terminal with a rich interactive TUI.

## Installation

### Build from source

```bash
go build -o shoehorn.exe ./cmd/shoehorn
```

### Homebrew (macOS / Linux)

```bash
brew tap shoehorn-dev/tap
brew install shoehorn
```

### Mise (macOS / Linux / Windows)

Add to your `mise.toml` or `~/.config/mise/config.toml`:

```toml
[tools."github:shoehorn-dev/cli"]
version = "latest"
exe = "shoehorn"
```

Then run:

```bash
mise install
```

**Windows:** Ensure the mise shims directory is in your PATH:

```powershell
# Add to your PowerShell profile (one-time setup):
Add-Content $PROFILE '$env:PATH = "$env:LOCALAPPDATA\mise\shims;$env:PATH"'
```

### Manual download

Download the binary for your platform from [Releases](https://github.com/shoehorn-dev/cli/releases), extract it, and add to your PATH.

---

## Quick Start

### 1. Authenticate with a Personal Access Token (recommended)

```bash
shoehorn auth login --server http://localhost:8080 --token shp_your_token_here
```

On success you'll see a panel like:

```
╭─ Authenticated with PAT ───────────────────────────────╮
│ ✓ Authenticated with PAT                               │
│                                                        │
│ Name      Jane Smith                                   │
│ Email     jane@example.com                             │
│ Tenant    acme-corp                                    │
│ Server    http://localhost:8080                        │
╰────────────────────────────────────────────────────────╯
```

### 2. Verify your identity

```bash
shoehorn whoami
```

### 3. Explore the catalog

```bash
shoehorn get entities
shoehorn search "payment"
shoehorn get entity payment-service --scorecard
```

---

## Authentication

### PAT login

```bash
shoehorn auth login --server http://localhost:8080 --token shp_xxxx
```

### Check auth status

```bash
shoehorn auth status
```

```
Profile: default
Server:  http://localhost:8080
Status:  Authenticated (PAT)
Email:   jane@example.com
Tenant:  acme-corp
Token:   Valid (PAT, no expiry)
Server:  Token verified with server
```

### Logout

```bash
shoehorn auth logout
```

---

## Commands

### `whoami`

Show your full user profile including roles, groups, and teams.

```bash
shoehorn whoami
shoehorn whoami --output json
```

---

### `search`

Full-text search across all catalog entities. Results open in an interactive table — press `Enter` to expand any item.

```bash
shoehorn search "payment"
shoehorn search "kafka" --output json
```

---

### `get entities`

List all catalog entities in an interactive table.

```bash
shoehorn get entities
shoehorn get entities --type service
shoehorn get entities --owner platform-team
shoehorn get entities --output json
```

Flags:
- `--type` — filter by entity type (service, library, website, etc.)
- `--owner` — filter by owning team slug

---

### `get entity`

Full detail panel for a single entity.

```bash
shoehorn get entity payment-service
shoehorn get entity payment-service --scorecard
shoehorn get entity <uuid> --output json
```

Example output:

```
╭─ payment-service ──────────────────────────────────────────╮
│ payment-service                                            │
│                                                            │
│ Type               service                                 │
│ Owner              platform-team                           │
│ Lifecycle          production                              │
│ Tier               1                                       │
│ Description        Core payment processing service         │
│ Tags               payments  core  pci                     │
│ Status             ● healthy  (99.97% uptime)              │
│ Links              GitHub  Grafana  Datadog                │
│                                                            │
│ Resources (3)                                              │
│ payment-db         PostgreSQL  production                  │
│ payment-cache      Redis       production                  │
│ payment-queue      Kafka topic production                  │
│                                                            │
│ Scorecard                                                  │
│ Grade              A  ████████████████████████░░░░ 92/100  │
╰────────────────────────────────────────────────────────────╯
```

---

### `get teams`

List all teams.

```bash
shoehorn get teams
shoehorn get teams --output json
```

---

### `get team`

Full detail for a team, including its members.

```bash
shoehorn get team platform-team
shoehorn get team <uuid>
```

---

### `get users`

List all users in the directory.

```bash
shoehorn get users
shoehorn get users --output json
```

---

### `get user`

Detail for a specific user: groups, teams, roles.

```bash
shoehorn get user <user-id>
```

---

### `get groups`

List all directory groups.

```bash
shoehorn get groups
```

---

### `get group`

Show roles mapped to a specific group.

```bash
shoehorn get group engineers
```

---

### `get owned`

List all entities owned by a specific team or user.

```bash
shoehorn get owned --by team platform-team
shoehorn get owned --by user <user-id>
```

---

### `get scorecard`

Scorecard breakdown for an entity with a visual score bar and per-check table.

```bash
shoehorn get scorecard payment-service
shoehorn get scorecard payment-service --output json
```

---

### `get k8s`

List all registered Kubernetes agents.

```bash
shoehorn get k8s
shoehorn get k8s --output json
```

---

### `forge molds list`

List all available Forge workflow templates.

```bash
shoehorn forge molds list
shoehorn forge molds list --output json
```

---

### `forge molds get`

Detail view for a mold: actions, inputs, and steps.

```bash
shoehorn forge molds get create-empty-github-repo
```

---

### `forge execute`

Execute a mold workflow in one step. Fetches the mold, resolves the action, fills defaults, validates required inputs, and creates a run.

```bash
shoehorn forge execute create-empty-github-repo \
  --input name=my-service \
  --input owner=my-org

# Specify a non-primary action
shoehorn forge execute my-mold --action scaffold --input name=my-app

# Validate without executing
shoehorn forge execute create-empty-github-repo \
  --input name=test --input owner=my-org --dry-run

# Pass inputs as JSON
shoehorn forge execute my-mold --inputs '{"name":"my-svc","owner":"my-org"}'
```

Flags:
- `--input` — repeatable `key=value` pairs (types are coerced from the mold schema)
- `--inputs` — JSON object with all inputs
- `--action` — action name (auto-selects primary action if omitted)
- `--dry-run` — validate without executing

---

### `forge run list`

List all workflow runs.

```bash
shoehorn forge run list
shoehorn forge run list --output json
```

---

### `forge run get`

Detail for a specific run.

```bash
shoehorn forge run get <run-id>
shoehorn forge run get <run-id> --output json
```

---

### `forge run create`

Start a new workflow run from a mold (lower-level than `forge execute`).

```bash
shoehorn forge run create create-empty-github-repo --action create \
  --input name=my-service --input owner=my-org
```

---

## Global Flags

All commands accept these flags:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | (interactive) | Output format: `json`, `yaml`, or `table` |
| `--no-interactive` | `-I` | `false` | Disable TUI, force plain text output |
| `--profile` | | `default` | Auth profile to use |
| `--config` | | `~/.shoehorn/config.yaml` | Config file path |

### Script-friendly output

Any command can be piped to `jq` or used in scripts:

```bash
shoehorn get entities --output json | jq '.[] | select(.type == "service") | .name'
shoehorn get team platform-team --output json | jq '.members[].email'
shoehorn whoami --output json | jq '.tenant_id'
```

---

## Interactive TUI Controls

All table views share the same key bindings:

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Select item / expand details |
| `q` / `Esc` | Quit / clear filter |
| `Backspace` | Remove last filter character |

---

## Configuration

Config file: `~/.shoehorn/config.yaml`

```yaml
version: "1.0"
current_profile: default

profiles:
  default:
    name: Default
    server: http://localhost:8080
    auth:
      provider_type: pat
      access_token: shp_xxxxxxxxxxxx
      user:
        email: jane@example.com
        name: Jane Smith
        tenant_id: acme-corp

  prod:
    name: Production
    server: https://api.shoehorn.dev
    auth:
      provider_type: pat
      access_token: shp_prod_xxxx
      user:
        email: jane@example.com
        tenant_id: acme-corp
```

### Multiple profiles

```bash
# Login to each environment
shoehorn --profile prod auth login --server https://api.shoehorn.dev --token shp_prod_xxx
shoehorn --profile staging auth login --server https://staging.shoehorn.dev --token shp_stg_xxx

# Use a specific profile for any command
shoehorn --profile prod get entities
shoehorn --profile staging forge molds list
```

---

## Project Structure

```
cli/
├── cmd/shoehorn/
│   ├── main.go                    # Entry point
│   └── commands/
│       ├── root.go                # Root command + global flags
│       ├── auth.go                # auth login/status/logout
│       ├── whoami.go              # whoami
│       ├── search.go              # search <query>
│       ├── forge.go               # forge run/molds
│       └── get/
│           ├── get.go             # get (parent command)
│           ├── entities.go        # get entities / get entity
│           ├── teams.go           # get teams / get team
│           ├── users.go           # get users / get user
│           ├── groups.go          # get groups / get group
│           ├── owned.go           # get owned
│           ├── scorecard.go       # get scorecard
│           └── k8s.go             # get k8s
├── pkg/
│   ├── api/
│   │   ├── client.go              # HTTP client + NewClientFromConfig
│   │   ├── auth.go                # Device flow types + methods
│   │   ├── catalog.go             # Catalog API: entities, teams, users, forge...
│   │   └── manifests.go           # Manifest types
│   ├── config/
│   │   └── config.go              # Config file, profiles, PAT helpers
│   ├── tui/
│   │   ├── styles.go              # Shared lipgloss styles
│   │   ├── spinner.go             # RunSpinner() helper
│   │   ├── table.go               # RunTable() interactive table
│   │   └── detail.go              # RenderDetail(), score bars, boxes
│   └── ui/
│       ├── detect.go              # Interactive vs plain mode detection
│       └── output.go              # JSON/YAML rendering
└── go.mod
```

---

## Troubleshooting

### "not authenticated" error

```bash
shoehorn auth status
shoehorn auth login --server http://localhost:8080 --token shp_xxxx
```

### API connection refused

Check that the Shoehorn API is running:

```bash
curl http://localhost:8080/health
```

Verify the server URL in your config:

```bash
cat ~/.shoehorn/config.yaml
```

### Token rejected by server

Your PAT may have been revoked. Generate a new one in the Shoehorn UI and re-authenticate:

```bash
shoehorn auth logout
shoehorn auth login --server http://localhost:8080 --token shp_new_token
```

### TUI not rendering correctly

Disable interactive mode if your terminal doesn't support ANSI colors:

```bash
shoehorn get entities --no-interactive
shoehorn get entities -o json
```
