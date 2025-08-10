# Git Branching Strategy and Semantic Versioning Framework

## Overview

This project follows **GitLab Flow with Release Branches** combined with **Semantic Versioning (SemVer)** to ensure reliable, predictable releases while maintaining development velocity.

## Branching Strategy: GitLab Flow with Release Branches

### Core Principles

1. **main branch**: Integration branch for ongoing development
2. **Release branches**: Long-lived branches for supported versions (e.g., `release/v1.2`, `release/v2.0`)
3. **Feature branches**: Short-lived branches for individual features/fixes
4. **Upstream-first hotfixes**: All fixes start from main and flow downstream

### Branch Types

#### main Branch
- **Purpose**: Primary integration branch containing the latest development
- **Protection**: Protected branch requiring PR reviews
- **Deployment**: Can be deployed to development/testing environments
- **Versioning**: Pre-release versions (e.g., `v1.3.0-alpha.1`, `v1.3.0-beta.2`)

#### Release Branches (e.g., `release/v1.2`, `release/v2.0`)
- **Purpose**: Stable branches for specific major.minor versions
- **Creation**: Created from main when ready to start a new release line
- **Updates**: Only critical bug fixes and security patches
- **Versioning**: Patch versions only (e.g., `v1.2.0`, `v1.2.1`, `v1.2.2`)
- **Support**: Multiple release branches can be maintained simultaneously

#### Feature Branches (e.g., `feature/network-analysis`, `fix/memory-leak`)
- **Purpose**: Development of new features or bug fixes
- **Origin**: Always created from main
- **Naming**: Use conventional prefixes:
  - `feat/` - New features
  - `fix/` - Bug fixes
  - `docs/` - Documentation updates
  - `refactor/` - Code refactoring
  - `test/` - Test improvements
- **Lifecycle**: Merged back to main via Pull Request, then deleted

#### Hotfix Workflow (Upstream-First)
1. Create hotfix branch from main (not from release branch)
2. Implement and test the fix
3. Merge to main first (creates new pre-release version)
4. Cherry-pick or merge to affected release branches
5. Tag patch versions on release branches

## Semantic Versioning (SemVer)

### Version Format: MAJOR.MINOR.PATCH

- **MAJOR** (X.y.z): Breaking changes to public API
- **MINOR** (x.Y.z): New backward-compatible features
- **PATCH** (x.y.Z): Backward-compatible bug fixes

### Pre-release Versions

- **Alpha** (`v1.3.0-alpha.1`): Early development, main branch
- **Beta** (`v1.3.0-beta.1`): Feature-complete, testing phase
- **Release Candidate** (`v1.3.0-rc.1`): Stable, final testing

### Version Precedence

```
1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-beta < 1.0.0-beta.2 < 1.0.0-rc.1 < 1.0.0
```

## Conventional Commits

We use [Conventional Commits](https://www.conventionalcommits.org/) to automate version bumping and changelog generation.

### Commit Message Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types and Version Impact

- `feat:` → MINOR version bump (new feature)
- `fix:` → PATCH version bump (bug fix)
- `BREAKING CHANGE:` footer → MAJOR version bump
- `docs:`, `style:`, `refactor:`, `test:`, `chore:` → No version bump

### Examples

```bash
# PATCH bump (1.2.0 → 1.2.1)
git commit -m "fix(api): resolve memory leak in client connection pool"

# MINOR bump (1.2.1 → 1.3.0)
git commit -m "feat(network): add deep packet inspection analysis"

# MAJOR bump (1.3.0 → 2.0.0)
git commit -m "feat(api): redesign configuration interface

BREAKING CHANGE: Configuration structure has changed. See migration guide."
```

## Workflow Examples

### Feature Development

```bash
# Start new feature
git checkout main
git pull origin main
git checkout -b feat/alert-system

# Development with conventional commits
git commit -m "feat(alerts): add alert manager interface"
git commit -m "feat(alerts): implement email notifications"
git commit -m "test(alerts): add comprehensive test coverage"

# Create Pull Request to main
gh pr create --title "feat: implement alert system" --body "Adds configurable alerts with email notifications"

# After review and merge, branch is automatically deleted
```

### Release Creation

```bash
# Create new release branch from main
git checkout main
git pull origin main
git checkout -b release/v1.3

# Stabilization commits (bug fixes only)
git commit -m "fix(build): resolve Docker build issue"
git commit -m "docs(api): update configuration examples"

# Tag initial release
git tag v1.3.0
git push origin release/v1.3 --tags

# Continue patch releases on this branch
git commit -m "fix(security): patch vulnerability in dependency"
git tag v1.3.1
git push origin release/v1.3 --tags
```

### Hotfix Process (Upstream-First)

```bash
# Critical bug in production v1.2.5
# 1. Fix on main first
git checkout main
git checkout -b fix/critical-security-issue
git commit -m "fix(security): patch critical authentication bypass"

# Merge to main
gh pr create --title "fix: critical security patch"
# After merge, this creates v1.4.0-alpha.1

# 2. Cherry-pick to release branch
git checkout release/v1.2
git cherry-pick <commit-hash>
git tag v1.2.6
git push origin release/v1.2 --tags
```

## Automation

### Automated Versioning

- **semantic-release** analyzes conventional commits
- **Automatic tagging** on main and release branches
- **Changelog generation** from commit history
- **GitHub releases** with artifacts

### CI/CD Integration

1. **PR to main**: Run tests, build verification
2. **Merge to main**: Create pre-release version, deploy to staging
3. **Push to release branch**: Create stable release, deploy to production
4. **Tag creation**: Trigger release artifacts, publish to registries

## Branch Protection Rules

### main branch
- Require PR reviews (2 reviewers)
- Require status checks (CI, tests)
- Require up-to-date branches
- No direct pushes
- No force pushes

### release/* branches
- Require PR reviews (1 reviewer)
- Require status checks
- Allow hotfix commits from maintainers
- No force pushes

## Migration Guide

### From Current State

1. **Protect main branch** with required reviews
2. **Install semantic-release** and configure automation
3. **Adopt conventional commits** for all new work
4. **Create first release branch** when ready for v1.0.0
5. **Update CI/CD pipelines** for automated releases

### Team Training

- Review conventional commit standards
- Practice feature branch workflow
- Understand upstream-first hotfix process
- Learn semantic versioning principles

## References

- [Semantic Versioning Specification](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [GitLab Flow Documentation](https://docs.gitlab.com/ee/topics/gitlab_flow.html)
- [semantic-release Documentation](https://semantic-release.gitbook.io/)
