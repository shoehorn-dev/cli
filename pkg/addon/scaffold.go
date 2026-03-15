// Package addon provides utilities for addon development (scaffold, build, publish).
package addon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// Tier represents the addon capability tier.
type Tier string

const (
	TierDeclarative Tier = "declarative"
	TierScripted    Tier = "scripted"
	TierFull        Tier = "full"
)

// ValidTiers is the set of valid tiers.
var ValidTiers = map[Tier]bool{
	TierDeclarative: true,
	TierScripted:    true,
	TierFull:        true,
}

// slugRegexp validates addon slugs (kebab-case, 3-50 chars).
var slugRegexp = regexp.MustCompile(`^[a-z][a-z0-9-]{1,48}[a-z0-9]$`)

// ScaffoldConfig holds the configuration for scaffolding a new addon.
type ScaffoldConfig struct {
	Name string // Addon slug (kebab-case)
	Tier Tier   // declarative, scripted, full
	Dir  string // Output directory (defaults to Name)
}

// ValidateSlug checks if a slug is valid.
func ValidateSlug(slug string) error {
	if !slugRegexp.MatchString(slug) {
		return fmt.Errorf("invalid slug %q: must be kebab-case, 3-50 chars, start/end with letter/digit", slug)
	}
	return nil
}

// Scaffold creates a new addon project directory.
func Scaffold(cfg ScaffoldConfig) error {
	if err := ValidateSlug(cfg.Name); err != nil {
		return err
	}
	if !ValidTiers[cfg.Tier] {
		return fmt.Errorf("invalid tier %q: must be declarative, scripted, or full", cfg.Tier)
	}

	dir := cfg.Dir
	if dir == "" {
		dir = cfg.Name
	}

	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("directory %q already exists", dir)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	files := scaffoldFiles(cfg)
	for relPath, content := range files {
		fullPath := filepath.Join(dir, relPath)

		// Ensure parent directory exists
		if parent := filepath.Dir(fullPath); parent != dir {
			if err := os.MkdirAll(parent, 0755); err != nil {
				return fmt.Errorf("create directory %s: %w", parent, err)
			}
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("write %s: %w", relPath, err)
		}
	}

	return nil
}

// scaffoldFiles returns the map of relative path -> file content for the scaffold.
func scaffoldFiles(cfg ScaffoldConfig) map[string]string {
	data := templateData{
		Name:        cfg.Name,
		DisplayName: slugToDisplayName(cfg.Name),
		Tier:        string(cfg.Tier),
	}

	files := map[string]string{
		"manifest.json": renderTemplate(manifestTemplate, data),
		"README.md":     renderTemplate(readmeTemplate, data),
	}

	switch cfg.Tier {
	case TierDeclarative:
		// Declarative addons are YAML-only, no TypeScript
		// manifest.json is all that's needed

	case TierScripted, TierFull:
		files["package.json"] = renderTemplate(packageJSONTemplate, data)
		files["tsconfig.json"] = tsconfigContent
		files["esbuild.config.mjs"] = esbuildConfigContent
		files["src/index.ts"] = renderTemplate(indexTSTemplate, data)

		if cfg.Tier == TierFull {
			files["src/index.ts"] = renderTemplate(indexTSFullTemplate, data)
		}
	}

	return files
}

type templateData struct {
	Name        string
	DisplayName string
	Tier        string
}

func slugToDisplayName(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

func renderTemplate(tmplStr string, data templateData) string {
	tmpl := template.Must(template.New("").Parse(tmplStr))
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// GenerateManifestJSON creates a manifest.json from ScaffoldConfig.
func GenerateManifestJSON(cfg ScaffoldConfig) ([]byte, error) {
	manifest := map[string]interface{}{
		"schemaVersion": 1,
		"kind":          "addon",
		"metadata": map[string]interface{}{
			"slug":     cfg.Name,
			"name":     slugToDisplayName(cfg.Name),
			"version":  "0.1.0",
			"category": "custom",
			"tier":     "free",
		},
		"addon": map[string]interface{}{
			"tier":    string(cfg.Tier),
			"runtime": "quickjs",
		},
	}

	return json.MarshalIndent(manifest, "", "  ")
}

// ─── Templates ──────────────────────────────────────────────────────────────

var manifestTemplate = `{
  "schemaVersion": 1,
  "kind": "addon",
  "metadata": {
    "slug": "{{.Name}}",
    "name": "{{.DisplayName}}",
    "version": "0.1.0",
    "description": "A Shoehorn addon",
    "author": {
      "name": "Your Name"
    },
    "category": "custom",
    "tier": "free"
  },
  "addon": {
    "tier": "{{.Tier}}",
    "runtime": "quickjs",
    "permissions": {
      "network": [],
      "shoehorn": ["entities:read"]
    }
  }
}
`

var packageJSONTemplate = `{
  "name": "{{.Name}}",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "build": "node esbuild.config.mjs",
    "dev": "node esbuild.config.mjs --watch",
    "typecheck": "tsc --noEmit"
  },
  "devDependencies": {
    "esbuild": "^0.21.0",
    "typescript": "^5.5.0"
  }
}
`

var tsconfigContent = `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ES2020",
    "moduleResolution": "bundler",
    "strict": true,
    "outDir": "dist",
    "rootDir": "src",
    "declaration": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  },
  "include": ["src"]
}
`

var esbuildConfigContent = `import { build } from 'esbuild';

const isWatch = process.argv.includes('--watch');

// QuickJS requires IIFE format with global exports.
// globalName wraps exports; footer hoists them to global scope
// so the Shoehorn runtime can call handleRoute() directly.
const config = {
  entryPoints: ['src/index.ts'],
  bundle: true,
  outfile: 'dist/addon.js',
  format: 'iife',
  globalName: '__addon__',
  footer: { js: 'if(typeof __addon__!=="undefined"){for(var k in __addon__)globalThis[k]=__addon__[k];}' },
  target: 'es2020',
  platform: 'neutral',
  minify: !isWatch,
  sourcemap: isWatch,
};

if (isWatch) {
  const ctx = await build({ ...config, plugins: [] });
  console.log('Watching for changes...');
} else {
  await build(config);
  console.log('Build complete: dist/addon.js');
}
`

var indexTSTemplate = `/**
 * {{.DisplayName}} - Shoehorn Addon ({{.Tier}})
 *
 * Functions are called by the QuickJS runtime with JSON string arguments.
 * handleRoute receives the request as a JSON string and must return a JSON string.
 */

interface RouteRequest {
  method: string;
  path: string;
  headers?: Record<string, string>;
  query?: Record<string, string>;
  body?: string;
}

interface RouteResponse {
  status: number;
  body?: string;
  headers?: Record<string, string>;
}

/**
 * Handle incoming HTTP requests routed to this addon.
 * Called by runtime as: handleRoute('{"method":"GET","path":"/ping",...}')
 */
export function handleRoute(requestJSON: string): string {
  const request: RouteRequest = JSON.parse(requestJSON);

  if (request.path === '/ping') {
    const response: RouteResponse = {
      status: 200,
      body: JSON.stringify({ message: 'pong', addon: '{{.Name}}' }),
    };
    return JSON.stringify(response);
  }

  return JSON.stringify({ status: 404, body: JSON.stringify({ error: 'not found' }) });
}

/**
 * Sync function called on schedule (if configured in manifest).
 */
export function sync(): string {
  return JSON.stringify({ synced: 0 });
}
`

var indexTSFullTemplate = `/**
 * {{.DisplayName}} - Shoehorn Addon (full)
 *
 * Full-tier addon with access to external resources (Postgres, Kafka, etc.)
 * Functions are called by the QuickJS runtime with JSON string arguments.
 */

interface RouteRequest {
  method: string;
  path: string;
  headers?: Record<string, string>;
  query?: Record<string, string>;
  body?: string;
}

interface RouteResponse {
  status: number;
  body?: string;
  headers?: Record<string, string>;
}

/**
 * Handle incoming HTTP requests routed to this addon.
 * Called by runtime as: handleRoute('{"method":"GET","path":"/ping",...}')
 */
export function handleRoute(requestJSON: string): string {
  const request: RouteRequest = JSON.parse(requestJSON);

  if (request.path === '/ping') {
    const response: RouteResponse = {
      status: 200,
      body: JSON.stringify({ message: 'pong', addon: '{{.Name}}' }),
    };
    return JSON.stringify(response);
  }

  return JSON.stringify({ status: 404, body: JSON.stringify({ error: 'not found' }) });
}

/**
 * Sync function called on schedule (if configured in manifest).
 * For full-tier addons, use ctx.postgres.query() and ctx.entities.upsert()
 * via host functions (available at runtime, not at build time).
 */
export function sync(): string {
  // Example: host_postgres_query('SELECT datname FROM pg_database')
  // Then: host_entities_upsert(JSON.stringify({name: row.datname, type: 'database'}))
  return JSON.stringify({ synced: 0 });
}
`

var readmeTemplate = `# {{.DisplayName}}

A Shoehorn addon (tier: {{.Tier}}).

## Development

` + "```" + `bash
# Install dependencies (scripted/full only)
npm install

# Build the addon bundle
npm run build
# Or: shoehorn addon build

# Publish to your Shoehorn instance
shoehorn addon publish
` + "```" + `

## Structure

- ` + "`manifest.json`" + ` - Addon manifest (permissions, metadata, config)
- ` + "`src/index.ts`" + ` - Addon entry point (handleRequest, sync)
- ` + "`dist/addon.js`" + ` - Compiled bundle (generated by build)
`
