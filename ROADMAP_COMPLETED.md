---

## Phase 1: Foundation & Project Setup

Establish the project structure, tooling, and basic operator scaffolding.

### 1.1 Project Scaffolding
- [ ] Initialize Go module (`go mod init github.com/CraightonH/boilerr`)
- [ ] Set up Kubebuilder project structure
- [ ] Create initial Makefile with common targets (build, test, generate, deploy)
- [ ] Add `.gitignore` for Go/Kubernetes projects
- [ ] Configure linting (golangci-lint)
- [ ] Set up pre-commit hooks

### 1.2 CI/CD Pipeline
- [ ] GitHub Actions workflow for PR checks (lint, test, build)
- [ ] Container image build and push workflow
- [ ] Release workflow with semantic versioning
- [ ] CRD schema validation in CI

### 1.3 Documentation Foundation
- [ ] Set up contributing guidelines (CONTRIBUTING.md)
- [ ] Create issue and PR templates
- [ ] Add code of conduct
- [ ] License selection and LICENSE file
