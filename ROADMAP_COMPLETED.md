---

## Phase 1: Foundation & Project Setup

Establish the project structure, tooling, and basic operator scaffolding.

### 1.1 Project Scaffolding
- [x] Initialize Go module (`go mod init github.com/CraightonH/boilerr`)
- [x] Set up Kubebuilder project structure
- [x] Create initial Makefile with common targets (build, test, generate, deploy)
- [x] Add `.gitignore` for Go/Kubernetes projects
- [x] Configure linting (golangci-lint)
- [x] Set up pre-commit hooks

### 1.2 CI/CD Pipeline
- [x] GitHub Actions workflow for PR checks (lint, test, build)
- [x] Container image build and push workflow
- [x] Release workflow with semantic versioning
- [x] CRD schema validation in CI

### 1.3 Documentation Foundation
- [x] Set up contributing guidelines (CONTRIBUTING.md)
- [x] Create issue and PR templates
- [x] Add code of conduct
- [x] License selection and LICENSE file
