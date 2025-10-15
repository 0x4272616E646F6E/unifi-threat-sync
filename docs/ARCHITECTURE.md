# Project Structure

```
unifi-threat-sync/
├── cmd/
│   └── unifi-threat-sync/
│       └── main.go              # Application entry point
│
├── internal/
│   ├── config/
│   │   ├── config.go            # Config loading and validation
│   │   └── config_test.go
│   │
│   ├── parser/
│   │   ├── parser.go            # Parser interface definition
│   │   ├── registry.go          # Parser registration and factory
│   │   ├── plain.go             # Plain text parser
│   │   ├── plain_test.go
│   │   ├── netset.go            # FireHOL netset parser
│   │   ├── netset_test.go
│   │   ├── abuseipdb.go         # AbuseIPDB API parser
│   │   ├── abuseipdb_test.go
│   │   ├── alienvault.go        # AlienVault OTX parser
│   │   ├── alienvault_test.go
│   │   └── utils.go             # Shared parsing utilities
│   │
│   ├── fetcher/
│   │   ├── fetcher.go           # HTTP client with caching
│   │   ├── cache.go             # ETag and hash-based caching
│   │   └── fetcher_test.go
│   │
│   ├── normalizer/
│   │   ├── normalizer.go        # IP/CIDR validation and deduplication
│   │   ├── merger.go            # Merge multiple feeds
│   │   └── normalizer_test.go
│   │
│   ├── unifi/
│   │   ├── client.go            # UniFi API client
│   │   ├── auth.go              # UniFi authentication
│   │   ├── groups.go            # Address group management
│   │   ├── rules.go             # Firewall rule management
│   │   └── client_test.go
│   │
│   └── sync/
│       ├── sync.go              # Main sync orchestration
│       ├── diff.go            # Calculate diffs (what to add/remove)
│       └── sync_test.go
│
├── configs/
│   └── config.yaml              # Example configuration
│
├── deploy/
│   ├── docker-compose.yaml      # Docker Compose example
│   └── kubernetes/
│       ├── deployment.yaml
│       ├── configmap.yaml
│       └── secret.yaml
│
├── docs/
│   ├── PARSERS.md               # Parser architecture documentation
│   └── CONTRIBUTING.md          # Contribution guidelines
│
├── Dockerfile                   # Multi-stage distroless build
├── Makefile                     # Build and development tasks
├── go.mod
├── go.sum
├── README.md
├── LICENSE
└── .dockerignore
```

## Package Descriptions

### `cmd/unifi-threat-sync`
Application entry point. Handles:
- Command-line flags
- Config loading
- Graceful shutdown
- Signal handling

### `internal/config`
Configuration management:
- Load YAML config
- Environment variable interpolation
- Validation
- Default values

### `internal/parser`
Feed parsers:
- Parser interface and registry
- Individual parser implementations
- Authentication handling
- Format-specific parsing logic

### `internal/fetcher`
HTTP fetching with caching:
- HTTP client wrapper
- ETag support
- Content hashing
- Rate limiting
- Retry logic

### `internal/normalizer`
IP/CIDR processing:
- Validate IP addresses and CIDR blocks
- Deduplicate entries
- Merge overlapping ranges (optional)
- Sort and normalize format

### `internal/unifi`
UniFi API client:
- Authentication (login/cookie management)
- Address group CRUD operations
- Firewall rule CRUD operations
- Error handling and retries

### `internal/sync`
Sync orchestration:
- Main sync loop
- Fetch all feeds
- Normalize and merge
- Calculate diff (add/remove)
- Update UniFi
- Hash tracking for idempotency

## Key Workflows

### Startup
```
main.go
  ├─> Load config
  ├─> Validate config
  ├─> Initialize UniFi client
  ├─> Test UniFi connection
  └─> Start sync loop
```

### Sync Loop
```
sync.Run()
  ├─> For each enabled feed:
  │     ├─> Fetch (with caching)
  │     ├─> Parse (using registered parser)
  │     └─> Collect results
  │
  ├─> Normalize & deduplicate all IPs
  ├─> Calculate aggregate hash
  ├─> Compare with previous hash
  │     └─> If same, skip update
  │
  ├─> Fetch current UniFi group
  ├─> Calculate diff (add/remove)
  ├─> Update UniFi address group
  ├─> Ensure firewall rule exists
  └─> Store new hash
```

### Adding a New Parser
```
1. Create internal/parser/myparser.go
2. Implement Parser interface
3. Add to registry in init()
4. Write tests
5. Update docs
```

## Dependencies (go.mod)

```go
module github.com/0x4272616E646F6E/unifi-threat-sync

go 1.25.2

require (
    gopkg.in/yaml.v3 v3.0.1           // YAML parsing
    github.com/go-resty/resty/v2      // HTTP client
    github.com/robfig/cron/v3         // Cron-style scheduling (optional)
)
```

## Testing Strategy

- **Unit tests**: Each package has `_test.go` files
- **Integration tests**: Test full sync flow with mock UniFi API
- **Test fixtures**: Sample feed data in `testdata/` folders
- **Table-driven tests**: For parsers and normalizers
- **Coverage goal**: 80%+

## Configuration Precedence

1. Command-line flags (highest priority)
2. Environment variables
3. Config file
4. Default values (lowest priority)

Example:
```go
// UniFi URL from CLI flag
--unifi-url=https://...

// Or from environment
UNIFI_URL=https://...

// Or from config file
unifi:
  url: https://...

// Or default
// (none - required field)
```
