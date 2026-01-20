# Contributing to Boilerr

Thank you for your interest in contributing to Boilerr! This document provides guidelines and information for contributors.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/boilerr.git`
3. Add upstream remote: `git remote add upstream https://github.com/CraightonH/boilerr.git`
4. Create a feature branch: `git checkout -b feature/my-feature`

## Development Setup

### Prerequisites

- Go 1.21+
- Docker
- kubectl
- A Kubernetes cluster (kind, minikube, or remote)
- kubebuilder (installed automatically via Makefile)

### Building

```bash
# Install dependencies and generate code
make generate
make manifests

# Build the operator binary
make build

# Run tests
make test

# Run linting
make lint
```

### Running Locally

```bash
# Install CRDs to cluster
make install

# Run operator locally against cluster
make run
```

## Making Changes

### Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation changes
- `chore:` maintenance tasks
- `test:` adding or updating tests
- `refactor:` code changes that neither fix bugs nor add features

### Pull Requests

1. Ensure your branch is up to date with main
2. Run `make lint test` before submitting
3. Update documentation if needed
4. Fill out the PR template completely
5. Request review from maintainers

### Adding a New Game CRD

See [CLAUDE.md](CLAUDE.md) for detailed instructions on adding game-specific CRDs.

## Testing

- Unit tests: `make test`
- E2E tests: `make test-e2e` (requires a cluster)
- Integration tests use envtest (no cluster required)

## Questions?

Open an issue with the `question` label or start a discussion.
