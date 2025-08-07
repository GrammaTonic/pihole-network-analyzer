# Feature Branch Workflow Guide

This repository now supports feature branch development with controlled builds and releases.

## Workflow Overview

### 🌿 **Feature Branches** (Any branch except main/master)
- ✅ **Tests run**: Full test suite, formatting, and security checks
- ✅ **Build verification**: Confirms code compiles successfully  
- ❌ **No artifacts**: No binaries or release artifacts are created
- ⚡ **Fast feedback**: Quick verification that changes are merge-ready

### 🚀 **Main/Master Branch** (Production)
- ✅ **Full CI/CD**: All tests, security scans, and builds
- ✅ **Multi-platform builds**: Linux, Windows, and macOS binaries
- ✅ **Artifacts**: Versioned releases with checksums
- ✅ **Auto-versioning**: Date and commit-based version tags

## Development Workflow

### 1. Create Feature Branch
```bash
git checkout -b feature/your-feature-name
git push -u origin feature/your-feature-name
```

### 2. Develop and Test
- Make your changes
- Commit and push to your feature branch
- GitHub Actions will:
  - Run tests ✅
  - Check formatting ✅
  - Verify build compatibility ✅
  - **NOT create build artifacts** ❌

### 3. Create Pull Request
```bash
# Create PR to main branch via GitHub UI
# All checks must pass before merge
```

### 4. Merge to Main
```bash
git checkout main
git pull origin main
# Merge via GitHub UI or:
git merge feature/your-feature-name
git push origin main
```

### 5. Production Build (Automatic)
When code is pushed to `main` or `master`:
- All tests run ✅
- Security scans execute ✅
- Multi-platform binaries are built 🏗️
- Artifacts are uploaded with version tags 📦
- Releases are created for tags 🚀

## CI/CD Jobs

| Job | Runs On | Purpose |
|-----|---------|---------|
| `test` | All branches | Unit tests, formatting, go vet |
| `security` | All branches | Vulnerability scanning |
| `build-check` | Feature branches only | Verify compilation |
| `build` | main/master only | Create production artifacts |

## Branch Protection

To enforce this workflow, consider setting up branch protection rules:

1. Go to **Settings** → **Branches**
2. Add rule for `main` branch:
   - ✅ Require pull request reviews
   - ✅ Require status checks to pass
   - ✅ Require conversation resolution
   - ✅ Include administrators

## Version Management

Production builds use automatic versioning:
- Format: `vYYYY.MM.DD-{commit-hash}`
- Example: `v2025.08.07-a1b2c3d`
- Embedded in binary via build flags

## Artifact Storage

- **Feature branches**: No artifacts (saves storage)
- **Main branch**: 90-day retention
- **Tagged releases**: Permanent storage

## Commands

### Test Locally Before Push
```bash
# Run the same tests as CI
go build -o pihole-network-analyzer .
./pihole-network-analyzer --test
go vet ./...
gofmt -s -l .
```

### Create Release
```bash
# Tag a commit for release
git tag v1.0.0
git push origin v1.0.0
# GitHub Actions will create a GitHub Release
```

## Benefits

- 🚀 **Fast feature development**: Quick feedback without heavy builds
- 💾 **Storage efficient**: No unnecessary artifacts on feature branches  
- 🔒 **Quality assurance**: All code must pass tests before merge
- 📦 **Reliable releases**: Only main branch creates production builds
- 🎯 **Clear separation**: Development vs production environments
