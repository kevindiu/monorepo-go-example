# CI/CD Guide

This document describes the CI/CD pipelines and workflows for the monorepo-go-example project.

## Overview

The project uses GitHub Actions for automated testing, building, and deployment. We have multiple workflows for different purposes:

1. **CI/CD Pipeline** (`ci.yml`) - Main pipeline for testing and deployment
2. **Pull Request Validation** (`pr-validation.yml`) - Fast validation for PRs
3. **E2E Tests** (`e2e-tests.yml`) - End-to-end testing workflow
4. **Release** (`release.yml`) - Release automation and image publishing

## Workflows

### 1. CI/CD Pipeline (ci.yml)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` branch

**Jobs:**

#### Test Job
- Sets up PostgreSQL database for integration tests
- Runs linter with golangci-lint
- Executes unit tests (`make test-unit`)
- Executes integration tests (`make test-integration`)
- Generates coverage report
- Uploads coverage to Codecov
- Builds all services

#### Build and Push Job
- Only runs on push to `main` branch
- Builds Docker images for all services
- Pushes images to GitHub Container Registry (ghcr.io)
- Scans images for vulnerabilities with Trivy
- Tags images with commit SHA

#### Deploy to Staging
- Deploys to staging environment after successful tests
- Updates Kubernetes manifests with new image tags
- Applies manifests to staging namespace
- Waits for rollout completion

#### Deploy to Production
- Requires staging deployment success
- Deploys to production environment
- Protected by GitHub environment approval (if configured)

### 2. Pull Request Validation (pr-validation.yml)

**Triggers:**
- Pull requests to `main` or `develop` branches

**Features:**
- **Code Quality Checks:**
  - Formatting validation (`make fmt`)
  - Static analysis (`make vet`)
  - Linting with golangci-lint
  
- **Testing:**
  - Unit tests with race detection
  - Coverage threshold check (minimum 50%)
  
- **Security:**
  - Vulnerability scanning with govulncheck
  - Dependency review for high-severity issues
  
- **Proto Validation:**
  - Lint proto files with buf
  - Check for breaking changes against main branch

### 3. E2E Tests (e2e-tests.yml)

**Triggers:**
- Manual workflow dispatch
- Daily schedule (2 AM UTC)
- Pull requests affecting services or E2E tests

**Workflow:**
1. Sets up PostgreSQL database
2. Builds all services
3. Starts services in background
4. Waits for services to be ready
5. Runs E2E test suite (`make test-e2e`)
6. Uploads logs on failure
7. Cleans up services

### 4. Release (release.yml)

**Triggers:**
- Push of version tags (e.g., `v1.0.0`)

**Workflow:**
1. **Create Release:**
   - Runs full test suite
   - Builds release binaries for multiple platforms
   - Generates changelog from commits
   - Creates GitHub release with binaries

2. **Build and Push Images:**
   - Builds multi-platform Docker images (amd64, arm64)
   - Tags with semantic version and commit SHA
   - Signs images with cosign
   - Pushes to GitHub Container Registry

3. **Deploy Helm Chart:**
   - Packages Helm chart
   - Uploads to GitHub release

## Environment Variables

### Required Secrets

Configure these in GitHub repository settings:

```yaml
# Container Registry
GITHUB_TOKEN: Automatically provided by GitHub Actions

# Kubernetes (Optional - for deployment)
KUBE_CONFIG: Base64 encoded kubeconfig for cluster access

# Codecov (Optional - for coverage reporting)
CODECOV_TOKEN: Token for uploading coverage reports
```

### Environment Variables in CI

```yaml
# Go Configuration
GO_VERSION: "1.21"

# Container Registry
REGISTRY: "ghcr.io"
IMAGE_NAME: "${{ github.repository }}"

# Database (Test Environment)
DATABASE_HOST: "localhost"
DATABASE_PORT: "5432"
DATABASE_USER: "postgres"
DATABASE_PASSWORD: "postgres"
DATABASE_NAME: "monorepo_test"
DATABASE_SSL_MODE: "disable"
```

## Running Workflows Locally

### Act (GitHub Actions Locally)

Install [act](https://github.com/nektos/act):

```bash
# Install act
brew install act  # macOS
# or
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Run workflow
act -j test
```

### Docker Compose for Local Testing

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run tests
make test-integration

# Cleanup
docker-compose -f docker-compose.test.yml down
```

## Deployment Environments

### Staging

- **Namespace:** `staging`
- **Trigger:** Automatic on push to `main`
- **URL:** `https://staging.example.com` (configure as needed)
- **Protection:** None (auto-deploy)

### Production

- **Namespace:** `production`
- **Trigger:** Manual approval after staging deployment
- **URL:** `https://example.com` (configure as needed)
- **Protection:** Requires approval in GitHub environment settings

### Environment Setup

To configure GitHub environments:

1. Go to repository **Settings** → **Environments**
2. Create environments: `staging` and `production`
3. For production, add required reviewers
4. Add environment secrets if needed

## Release Process

### Creating a Release

1. **Update version:**
   ```bash
   # Update VERSION in Makefile or version files
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **Automated steps:**
   - Tests run automatically
   - Binaries are built for all platforms
   - Docker images are built and pushed
   - GitHub release is created
   - Helm chart is packaged

3. **Manual deployment (if needed):**
   ```bash
   # Update Helm chart version
   helm upgrade monorepo-go-example charts/monorepo-go-example \
     --set image.tag=v1.0.0 \
     --namespace production
   ```

### Version Tags

Follow [Semantic Versioning](https://semver.org/):

- **Major:** `v1.0.0` - Breaking changes
- **Minor:** `v1.1.0` - New features
- **Patch:** `v1.1.1` - Bug fixes
- **Pre-release:** `v1.0.0-rc.1`, `v1.0.0-beta.1`, `v1.0.0-alpha.1`

## Monitoring CI/CD

### GitHub Actions UI

- **View workflows:** Repository → Actions tab
- **Check runs:** Click on workflow run for details
- **Download artifacts:** Available in workflow run summary

### Notifications

Configure GitHub notifications:

1. Repository → Settings → Notifications
2. Enable notifications for:
   - Failed workflows
   - Required reviews
   - Deployment status

### Status Badges

Add to README.md:

```markdown
![CI/CD](https://github.com/kevindiu/monorepo-go-example/workflows/CI%2FCD%20Pipeline/badge.svg)
![Tests](https://github.com/kevindiu/monorepo-go-example/workflows/Pull%20Request%20Validation/badge.svg)
[![codecov](https://codecov.io/gh/kevindiu/monorepo-go-example/branch/main/graph/badge.svg)](https://codecov.io/gh/kevindiu/monorepo-go-example)
```

## Troubleshooting

### Common Issues

#### Tests Failing in CI but Passing Locally

- Check database connection settings
- Verify environment variables
- Ensure all dependencies are committed

#### Docker Build Failures

- Check Dockerfile syntax
- Verify build context
- Ensure all files are included (not in .dockerignore)

#### Deployment Failures

- Verify Kubernetes credentials
- Check namespace exists
- Ensure image pull secrets are configured

### Debug Mode

Enable debug logging in workflow:

```yaml
- name: Enable debug mode
  run: echo "ACTIONS_STEP_DEBUG=true" >> $GITHUB_ENV
```

## Best Practices

1. **Keep workflows fast:**
   - Use caching for dependencies
   - Run only necessary tests in PR validation
   - Parallelize independent jobs

2. **Security:**
   - Never commit secrets
   - Use GitHub secrets for sensitive data
   - Scan images for vulnerabilities
   - Sign container images

3. **Reliability:**
   - Add retry logic for flaky tests
   - Use appropriate timeouts
   - Clean up resources after tests

4. **Maintainability:**
   - Keep workflows DRY (use reusable workflows)
   - Document custom actions
   - Version pin actions for stability

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [golangci-lint GitHub Action](https://github.com/golangci/golangci-lint-action)
- [Codecov Action](https://github.com/codecov/codecov-action)
