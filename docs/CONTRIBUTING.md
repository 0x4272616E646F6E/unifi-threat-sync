# Contributing to UniFi Threat Sync

Thank you for considering contributing to UniFi Threat Sync! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites
- Go 1.25+ installed
- Docker (optional, for building containers)
- Make (optional, but recommended)
- A UniFi controller for testing (or use Docker: `docker run -d --name unifi -p 8443:8443 linuxserver/unifi-network-application`)

### Setting Up Development Environment

1. **Clone the repository**
```bash
git clone https://github.com/0x4272616E646F6E/unifi-threat-sync.git
cd unifi-threat-sync
```

2. **Install dependencies**
```bash
go mod download
```

3. **Build the project**
```bash
make build
```

4. **Run tests**
```bash
make test
```

## Development Workflow

### Project Structure
```
unifi-threat-sync/
├── cmd/unifi-threat-sync/  # Application entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── parser/             # Feed parsers
│   ├── normalizer/         # IP/CIDR normalization
│   ├── unifi/              # UniFi API client
│   └── sync/               # Sync orchestration
├── configs/                # Example configuration
├── docs/                   # Documentation
└── Makefile               # Build automation
```

### Making Changes

1. **Create a feature branch**
```bash
git checkout -b feature/your-feature-name
```

2. **Make your changes**
   - Follow Go best practices
   - Add tests for new functionality
   - Update documentation as needed

3. **Run checks**
```bash
make check  # Runs fmt, vet, lint, and test
```

4. **Commit your changes**
```bash
git add .
git commit -m "feat: add new feature"
```

Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test additions/changes
- `refactor:` - Code refactoring
- `chore:` - Maintenance tasks

5. **Push and create PR**
```bash
git push origin feature/your-feature-name
```

## Adding a New Parser

To add support for a new threat feed format:

1. **Create parser file**: `internal/parser/yourparser.go`

```go
package parser

import (
    "context"
    "net"
    "github.com/0x4272616E646F6E/unifi-threat-sync/internal/config"
)

type YourParser struct{}

func init() {
    Register(&YourParser{})
}

func (p *YourParser) Name() string {
    return "yourparser"
}

func (p *YourParser) Parse(ctx context.Context, feedConfig config.FeedConfig) ([]net.IPNet, error) {
    // Implementation here
}

func (p *YourParser) ValidateConfig(feedConfig config.FeedConfig) error {
    // Validation here
}
```

2. **Create tests**: `internal/parser/yourparser_test.go`

```go
package parser

import "testing"

func TestYourParser_Parse(t *testing.T) {
    // Test implementation
}
```

3. **Update documentation**:
   - Add parser to `README.md` parser table
   - Document any auth requirements
   - Add example configuration

4. **Test thoroughly**:
```bash
make test
make lint
```

## Testing Guidelines

### Unit Tests
- Test all public functions
- Use table-driven tests
- Mock external dependencies
- Aim for 80%+ coverage

Example:
```go
func TestParseIPOrCIDR(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    net.IPNet
        wantErr bool
    }{
        {"valid IP", "192.168.1.1", ...},
        {"invalid IP", "999.999.999.999", ...},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Tests
- Test with real UniFi controller (if available)
- Use Docker for UniFi controller
- Test full sync cycle

### Running Tests
```bash
make test              # Run all tests
make coverage          # Generate coverage report
go test -v ./...       # Verbose output
go test -run TestName  # Run specific test
```

## Code Style

### Go Standards
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Run `go vet` for static analysis
- Use `golangci-lint` for comprehensive linting

### Best Practices
- Keep functions small and focused
- Use meaningful variable names
- Add comments for exported functions
- Handle errors explicitly
- Use context for cancellation

### Formatting
```bash
make fmt  # Format all code
make vet  # Run go vet
make lint # Run golangci-lint
```

## Documentation

### Code Documentation
- Add godoc comments for all exported types and functions
- Include usage examples where appropriate
- Document parameters and return values

Example:
```go
// Parse fetches and parses a threat feed, returning a list of IP networks.
// It uses the provided context for cancellation and timeout.
//
// Example:
//   networks, err := parser.Parse(ctx, feedConfig)
//   if err != nil {
//       return err
//   }
func Parse(ctx context.Context, feedConfig FeedConfig) ([]net.IPNet, error) {
    // Implementation
}
```

### User Documentation
- Update `README.md` for user-facing changes
- Update `docs/PARSERS.md` for parser changes
- Update `docs/ARCHITECTURE.md` for architectural changes
- Include examples and screenshots

## Pull Request Process

1. **Ensure all checks pass**
   - Tests pass
   - Lint passes
   - Documentation updated

2. **PR Description**
   - Clearly describe the change
   - Reference any related issues
   - Include screenshots/examples if applicable

3. **Review Process**
   - Address reviewer feedback
   - Keep PR focused and small
   - Rebase on main if needed

4. **Merge**
   - Squash commits if appropriate
   - Ensure commit message follows conventions

## Security

### Reporting Vulnerabilities
- **Do not** open public issues for security vulnerabilities
- Email security concerns to: [security contact]
- Provide detailed description and reproduction steps

### Security Best Practices
- Never commit credentials or API keys
- Use environment variables for secrets
- Validate all input
- Handle sensitive data carefully
- Keep dependencies updated

## Performance

### Optimization Guidelines
- Profile before optimizing
- Focus on hot paths
- Use streaming for large feeds
- Implement caching where appropriate
- Set appropriate timeouts

### Benchmarking
```bash
go test -bench=. ./...
go test -benchmem -bench=. ./...
```

## Release Process

1. **Version Bump**
   - Update version in appropriate files
   - Update CHANGELOG.md

2. **Tag Release**
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

3. **Build Artifacts**
```bash
make build-all
```

4. **Docker Image**
```bash
make docker-build
make docker-push
```

## Getting Help

- **Documentation**: Check `docs/` directory
- **Issues**: Search existing issues
- **Discussions**: Use GitHub Discussions
- **Examples**: See `configs/` directory

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
