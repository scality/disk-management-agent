# Contributing to disk-management-agent

Thank you for your interest in contributing. This document explains the
development workflow, coding standards, and process for submitting changes.

## Code of Conduct

All participants are expected to treat each other with respect and
professionalism. Harassment, discrimination, and disruptive behavior will not be
tolerated. Be constructive in code reviews and discussions.

## Development Workflow

1. Fork the repository and clone your fork.
2. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/my-change
   ```
3. Make your changes, ensuring they follow the coding standards below.
4. Run the full validation suite before pushing:
   ```bash
   make manifests generate fmt vet lint test
   ```
5. Push your branch and open a pull request against `main`.

## Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| Go | 1.25+ | Build and test |
| Make | any | Build system |
| Docker | any | Container image builds |
| Kind | any | End-to-end tests |
| kubectl | any | Cluster interaction |

All other tools (`controller-gen`, `kustomize`, `setup-envtest`, `golangci-lint`)
are downloaded automatically by the Makefile into `./bin/`.

## Setting Up the Development Environment

```bash
# Clone and enter the repository
git clone https://github.com/scality/disk-management-agent.git
cd disk-management-agent

# Download Go dependencies
go mod download

# Generate manifests, deepcopy, format, and vet
make manifests generate fmt vet

# Run unit tests to confirm everything works
make test
```

## Coding Standards

### Clean Architecture

The project follows clean architecture under `pkg/`:

- **`pkg/domain/`** -- Core business entities. No external dependencies.
- **`pkg/service/`** -- Interface definitions (ports). Use cases depend on these.
- **`pkg/usecase/`** -- Application business logic. Must not import
  infrastructure packages.
- **`pkg/infrastructure/`** -- Adapters that implement service interfaces.

Dependency direction is always **inward**: infrastructure depends on use cases
and service interfaces, never the reverse.

### Interface Naming

Interfaces follow the `<Entity><Action>er` pattern:

```go
type PhysicalDriveDiscoverer interface { ... }
type DiscoveredDriveCacheReader interface { ... }
```

Composite interfaces (repositories, services) may use broader names but must
embed small, focused interfaces.

### Error Handling

Wrap all errors with context using `errors.Wrap` or `fmt.Errorf` with `%w`:

```go
return errors.Wrap(err, fmt.Sprintf("disk %s not accessible", diskID))
```

Include relevant identifiers. Avoid duplicating context already present in the
error chain.

### Interface Size

Interfaces are limited to 1-2 methods. Only composite interfaces may have more
through embedding.

### Linting

The project uses [golangci-lint](https://golangci-lint.run/) with the
configuration in `.golangci.yml`. Run it locally before pushing:

```bash
make lint
```

## Testing

### Unit Tests

Unit tests and controller tests (using envtest) run together:

```bash
make test
```

Coverage output is written to `cover.out`.

### End-to-End Tests

E2E tests require a Kind cluster. The Makefile manages the cluster lifecycle:

```bash
make test-e2e
```

This creates a Kind cluster named `disk-management-agent-test-e2e`, runs the
tests, and tears it down automatically.

### Writing Tests

- Controller tests use the
  [envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest)
  framework and Ginkgo/Gomega.
- Use case tests use [testify](https://github.com/stretchr/testify) with mock
  implementations of service interfaces.
- Test files live alongside the code they test (`_test.go` suffix).

## Pull Request Process

1. Ensure all CI checks pass (lint + tests).
2. Keep PRs focused: one logical change per PR.
3. Write a clear PR description explaining **what** changed and **why**.
4. At least one approval from a code owner is required before merging.
5. Squash-merge is preferred for a clean commit history.

### Branch Naming

Use descriptive prefixes:

- `feature/` -- New functionality
- `fix/` -- Bug fixes
- `improvement/` -- Refactoring or enhancements
- `docs/` -- Documentation changes

## Issue Reporting

### Bug Reports

When reporting a bug, include:

- Steps to reproduce
- Expected behavior
- Actual behavior
- Environment details (Go version, Kubernetes version, RAID controller type)
- Relevant logs or error messages

### Feature Requests

Describe the use case, the expected behavior, and why the existing functionality
does not cover it.

## Documentation

When making code changes, update relevant documentation:

- If you add or modify CRD fields, regenerate manifests with `make manifests`.
- If you add environment variables, update the configuration table in
  `README.md`.
- If you change architecture or add components, update the architecture diagram
  in `README.md`.

## License

By contributing to this project, you agree that your contributions will be
licensed under the [Apache License 2.0](LICENSE).
