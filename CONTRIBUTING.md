# Contributing to K8s RBACtory Backend

Thank you for your interest in contributing to the K8s RBACtory backend! We welcome contributions from the community and are grateful for your support in making this project better.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Pull Request Process](#pull-request-process)
- [Testing](#testing)
- [Documentation](#documentation)
- [Getting Help](#getting-help)

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Review it [here](./CODE_OF_CONDUCT.md).

## Getting Started

### Prerequisites

- Go 1.23 or higher
- Docker (for containerized development and testing)
- kubectl (for Kubernetes integration testing)
- Make
- Git
- A code editor with Go support (VS Code with Go extension recommended)

### Initial Setup

1. Fork the repository on GitHub
2. Clone your fork locally:

   ```bash
   git clone https://github.com/YOUR_USERNAME/k8s-rbactory-backend.git
   cd k8s-rbactory-backend
   ```

3. Add the upstream repository:

   ```bash
   git remote add upstream https://github.com/djryanj/k8s-rbactory-backend.git
   ```

4. Install dependencies:

   ```bash
   make deps
   ```

5. Install development tools:

   ```bash
   make install-tools
   ```

6. Verify your setup:

   ```bash
   make verify-all
   ```

7. Create a branch for your work:
   ```bash
   git checkout -b feature/your-feature-name
   ```

### Running the Development Server

```bash
# Build and run
make dev

# Or run directly without building
make run
```

The API server will be available at `http://localhost:8080`.

### Available Make Targets

View all available commands:

```bash
make help
```

Common targets:

- `make build` - Build the binary
- `make test` - Run all tests
- `make test-coverage` - Run tests with coverage report
- `make lint` - Run linter
- `make fmt` - Format code
- `make clean` - Clean build artifacts

## Development Workflow

### Branching Strategy

We follow a simplified Git workflow:

- `main` - Production-ready code
- `develop` - Integration branch for features (if applicable)
- `feature/*` - New features
- `fix/*` - Bug fixes
- `docs/*` - Documentation updates
- `refactor/*` - Code refactoring
- `test/*` - Test additions or modifications
- `perf/*` - Performance improvements

### Keeping Your Fork Updated

Regularly sync your fork with the upstream repository:

```bash
git fetch upstream
git checkout main
git merge upstream/main
git push origin main
```

## Coding Standards

### Go Code Style

We follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go) guidelines.

#### General Principles

- Write idiomatic Go code
- Keep functions small and focused
- Use meaningful variable and function names
- Avoid premature optimization
- Write self-documenting code
- Add comments for complex logic

#### Formatting

All code must be formatted using `gofmt`:

```bash
# Format all code
make fmt

# Check formatting
gofmt -l .
```

#### Linting

We use `golangci-lint` for comprehensive code analysis:

```bash
# Run linter
make lint

# Run specific linters
golangci-lint run --enable=gofmt,govet,staticcheck
```

### Code Organization

```
backend/
├── cmd/
│   └── api/            # Application entry points
├── internal/
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # HTTP middleware
│   ├── k8s/            # Kubernetes client logic
│   ├── logging/        # Logging utilities
│   ├── models/         # Data models
│   └── testutil/       # Test utilities
├── pkg/                # Public libraries (if any)
├── test/               # Integration tests
├── api/                # API specifications (OpenAPI/Swagger)
└── hack/
    └── k8s-manifests/  # Sample kubernetes manfiests for deployment
```

### Naming Conventions

#### Packages

- Use short, lowercase, single-word names
- Avoid underscores or mixed caps
- Examples: `handlers`, `middleware`, `k8s`

#### Files

- Use lowercase with underscores for multi-word names
- Test files: `*_test.go`
- Examples: `server.go`, `role_handler.go`, `role_handler_test.go`

#### Variables and Functions

- Use camelCase for unexported names
- Use PascalCase for exported names
- Use short names for short-lived variables (e.g., `i`, `err`)
- Use descriptive names for longer-lived variables

```go
// Good
func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) error {
    var req CreateRoleRequest
    // ...
}

// Avoid
func (h *Handler) create_role(w http.ResponseWriter, r *http.Request) error {
    var createRoleRequest CreateRoleRequest
    // ...
}
```

#### Constants

- Use PascalCase for exported constants
- Use camelCase for unexported constants
- Group related constants in blocks

```go
const (
    DefaultPort     = 8080
    DefaultTimeout  = 30 * time.Second
    MaxRetries      = 3
)
```

### Error Handling

- Always check errors
- Provide context when wrapping errors
- Use `fmt.Errorf` with `%w` for error wrapping
- Return errors rather than panicking

```go
// Good
if err := doSomething(); err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Avoid
if err := doSomething(); err != nil {
    panic(err)
}
```

### Interfaces

- Keep interfaces small and focused
- Define interfaces where they are used, not where they are implemented
- Use standard interface names when appropriate (e.g., `Reader`, `Writer`)

```go
// Good - defined where used
type RoleGetter interface {
    GetRole(ctx context.Context, name string) (*Role, error)
}

// Avoid - overly broad interface
type RoleManager interface {
    GetRole(ctx context.Context, name string) (*Role, error)
    CreateRole(ctx context.Context, role *Role) error
    UpdateRole(ctx context.Context, role *Role) error
    DeleteRole(ctx context.Context, name string) error
    ListRoles(ctx context.Context) ([]*Role, error)
}
```

### Context Usage

- Always pass `context.Context` as the first parameter
- Never store context in a struct
- Use `context.Background()` only in main, init, and tests
- Propagate context through the call stack

```go
func (h *Handler) GetRole(ctx context.Context, name string) (*Role, error) {
    return h.k8sClient.GetRole(ctx, name)
}
```

## Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification for commit messages. This leads to more readable messages and enables automated changelog generation. Review the documentation at that site for more information.

### Scope

The scope should specify the area of the codebase affected:

- `handlers` - HTTP handlers
- `middleware` - HTTP middleware
- `k8s` - Kubernetes client logic
- `api` - API definitions
- `logging` - Logging functionality
- `server` - Server configuration
- `deps` - Dependency updates
- `build` - Build system changes

### Examples

```
feat(handlers): add support for ClusterRole creation

Implements ClusterRole creation endpoint with validation
and proper error handling. Includes support for aggregation
rules and non-resource URLs.

Closes #123
```

```
fix(k8s): resolve client timeout issue with large clusters

The Kubernetes client was timing out when listing resources
in clusters with more than 1000 namespaces. This fix increases
the default timeout and implements proper pagination.

Fixes #456
```

```
perf(handlers): optimize role listing with caching

Implement in-memory caching for role listings to reduce
Kubernetes API calls. Cache is invalidated on write operations.

Improves response time by 60% for repeated queries.
```

```
refactor(middleware): simplify authentication logic

Extract authentication logic into separate functions for
better testability and maintainability. No functional changes.
```

## Pull Request Process

**IMPORTANT**: PRs should only contain a single atomic change. For example, if a Go dependency update is required to enable a new feature, then the following fictional example PRs should be generated:

1. deps: update Go to 1.25
1. fix(k8s): resolve client timeout issue with large clusters

### Before Submitting

1. Ensure your branch is up to date with the main branch
1. Ensure your code follows the coding standards
1. Run the pre-commit checks:
   ```bash
   make pre-commit
   ```
1. Run the full test suite:
   ```bash
   make test-coverage
   ```
1. Verify there are no race conditions:
   ```bash
   make test-race
   ```
1. Update documentation if necessary

### PR Title

Use the same format as commit messages:

```
<type>(<scope>): <description>
```

Examples:

- `feat(handlers): add namespace filtering for role listings`
- `fix(k8s): correct RBAC policy generation for custom resources`
- `docs(api): update OpenAPI specification for new endpoints`

### PR Template

This project has a PR template that must be used. If PR's deviate from the PR template significantly they will be closed.

### Squash Commits

This project will always use squash commits when merging to main.

### Review Process

1. At least one maintainer must review and approve the PR
2. All CI checks must pass
3. Code coverage must not decrease
4. Address any feedback from reviewers
5. Once approved, a maintainer will merge your PR (squash commit)

### After Your PR is Merged

1. Delete your feature branch
2. Update your local repository:
   ```bash
   git checkout main
   git pull upstream main
   ```

## Testing

### Test Organization

Tests are organized by type:

- **Unit tests**: In the same package as the code (`*_test.go`)
- **Integration tests**: In the `test/` directory
- **Benchmarks**: In `*_test.go` files with `Benchmark` prefix

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run specific package tests
make test-handlers
make test-middleware
make test-k8s

# Run tests with coverage
make test-coverage

# Run tests with race detector
make test-race

# Run benchmarks
make test-bench

# Run a specific test
make test-one
# Then enter the test name when prompted
```

### Writing Tests

#### Unit Tests

```go
func TestRoleHandler_CreateRole(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateRoleRequest
        want    *Role
        wantErr bool
    }{
        {
            name: "valid role creation",
            input: CreateRoleRequest{
                Name: "test-role",
                Rules: []PolicyRule{
                    {
                        APIGroups: []string{""},
                        Resources: []string{"pods"},
                        Verbs:     []string{"get", "list"},
                    },
                },
            },
            want: &Role{
                Name: "test-role",
                // ...
            },
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### Integration Tests

```go
// +build integration

func TestRoleIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Integration test implementation
}
```

### Test Coverage Requirements

- Maintain or improve overall test coverage
- Aim for at least 80% coverage for new code
- Critical paths should have 100% coverage
- All exported functions should be tested

View coverage:

```bash
make test-coverage
# Opens coverage.html in your browser
```

### Testing Best Practices

1. **Table-Driven Tests**: Use table-driven tests for multiple scenarios
2. **Test Names**: Use descriptive test names that explain what is being tested
3. **Isolation**: Tests should be independent and not rely on execution order
4. **Cleanup**: Always clean up resources in tests
5. **Mocking**: Use interfaces for dependencies to enable mocking
6. **Error Cases**: Test both success and error paths
7. **Edge Cases**: Include tests for boundary conditions

### Race Detection

Always run tests with the race detector before submitting:

```bash
make test-race
```

## Documentation

### Code Documentation

- Add package documentation at the top of package files
- Document all exported functions, types, and constants
- Use complete sentences in comments
- Start comments with the name of the thing being documented

```go
// Package handlers provides HTTP request handlers for the API.
package handlers

// RoleHandler handles HTTP requests related to Kubernetes Roles.
type RoleHandler struct {
    client k8s.Client
}

// CreateRole creates a new Kubernetes Role in the specified namespace.
// It validates the role specification and returns an error if validation fails.
func (h *RoleHandler) CreateRole(ctx context.Context, req CreateRoleRequest) (*Role, error) {
    // Implementation
}
```

### API Documentation

- Update OpenAPI/Swagger specification for API changes
- Include request/response examples
- Document error responses
- Validate the specification:
  ```bash
  make swagger-validate
  ```

### README Updates

Update the README.md if your changes:

- Add new features or endpoints
- Change installation or setup procedures
- Modify configuration options
- Affect deployment

## Getting Help

### Communication Channels

- **GitHub Issues**: For bug reports and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Pull Request Comments**: For code review and implementation questions

### Asking Questions

When asking for help:

1. Search existing issues and discussions first
2. Provide context and relevant details
3. Include code samples or error messages
4. Describe what you have already tried
5. Mention your Go version and operating system
6. Be respectful and patient

### Reporting Bugs

When reporting bugs, include:

1. A clear, descriptive title
2. Steps to reproduce the issue
3. Expected behavior
4. Actual behavior
5. Environment details (Go version, OS, Kubernetes version)
6. Relevant logs or error messages
7. Minimal reproducible example if possible

Use the bug report template when creating an issue.

### Suggesting Features

When suggesting features:

1. Check if the feature has already been requested
2. Clearly describe the feature and its benefits
3. Provide use cases
4. Consider implementation complexity
5. Discuss potential API design
6. Be open to discussion and alternatives

Use the feature request template when creating an issue.

## Security

### Security Best Practices

- Never commit secrets or credentials
- Validate all user input
- Use parameterized queries
- Follow the principle of least privilege
- Keep dependencies up to date
- Handle errors securely (don't leak sensitive information)

### Reporting Security Issues

Please report security vulnerabilities privately to the maintainers rather than opening a public issue.

## Dependency Management

### Adding Dependencies

1. Add the dependency:

   ```bash
   go get github.com/example/package
   ```

2. Tidy modules:

   ```bash
   make mod-tidy
   ```

3. Verify dependencies:

   ```bash
   make verify-deps
   ```

4. Commit both `go.mod` and `go.sum`

### Updating Dependencies

```bash
# Update all dependencies
go get -u ./...
make mod-tidy

# Update specific dependency
go get -u github.com/example/package
make mod-tidy
```

## Continuous Integration

### CI Checks

All PRs must pass the following checks:

1. **Linting**: Code must pass `golangci-lint`
2. **Tests**: All tests must pass
3. **Race Detection**: No race conditions detected
4. **Coverage**: Coverage must not decrease
5. **Build**: Code must build successfully

Run all CI checks locally:

```bash
make ci
```

### Pre-commit Checks

Before committing, run:

```bash
make pre-commit
```

This runs:

- Code formatting
- Linting
- Unit tests
- Vet checks

## Release Process

Releases are handled by maintainers.

## Recognition

Contributors will be recognized in the following ways:

- Listed in the project's contributors list
- Mentioned in release notes for significant contributions

## License

By contributing to K8s RBACtory Backend, you agree that your contributions will be licensed under the Apache License 2.0, the same license as the project.

## Additional Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Conventional Commits](https://www.conventionalcommits.org/)

## Questions?

If you have questions about contributing that are not covered in this guide, please open a GitHub Discussion or reach out to the maintainers.

Thank you for contributing to K8s RBACtory Backend!

```

```
