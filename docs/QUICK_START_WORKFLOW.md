# Quick Start: Semantic Versioning & Branching Workflow

## 🚀 Initial Setup

### Node.js Requirement (One-Time)
semantic-release requires Node.js for release automation. This is a **development-only dependency** - your Go application has zero Node.js dependencies.

**Install Node.js:**
```bash
# macOS (recommended)
brew install node

# Or download from: https://nodejs.org/
```

**Setup semantic-release:**
```bash
# Check setup status
make release-status

# Install dependencies
make release-setup
```

### Configure Git (if not already done)
```bash
git config user.name "Your Name"
git config user.email "your.email@example.com"
```

## ✨ Why Node.js?
- **semantic-release** is the industry standard (30,000+ projects)
- **Complete automation**: commit → version → changelog → release
- **Zero manual release steps**
- **Node.js only used for CI/CD** - not in your Go application

## 📝 Daily Development Workflow

### 1. Start a New Feature

```bash
# Always start from main
git checkout main
git pull origin main

# Create feature branch
git checkout -b feat/your-feature-name
```

### 2. Make Commits (Using Conventional Format)

**Option A: Interactive commit (recommended for beginners)**
```bash
make commit
# Follow the prompts to create a properly formatted commit
```

**Option B: Manual commit**
```bash
git add .
git commit -m "feat(component): add new functionality"
```

### 3. Common Commit Types

| Type | Description | Version Impact | Example |
|------|-------------|----------------|---------|
| `feat` | New feature | MINOR bump | `feat(api): add user authentication` |
| `fix` | Bug fix | PATCH bump | `fix(network): resolve memory leak` |
| `docs` | Documentation | No bump | `docs: update API documentation` |
| `style` | Code style/formatting | No bump | `style: fix code formatting` |
| `refactor` | Code refactoring | PATCH bump | `refactor(analyzer): improve performance` |
| `test` | Tests | No bump | `test(api): add integration tests` |
| `chore` | Maintenance | No bump | `chore: update dependencies` |

### 4. Breaking Changes

For backward-incompatible changes (MAJOR version bump):

```bash
git commit -m "feat(api): redesign configuration interface

BREAKING CHANGE: Configuration structure has changed. 
See migration guide in docs/MIGRATION.md"
```

### 5. Create Pull Request

```bash
# Push your branch
git push origin feat/your-feature-name

# Create PR using GitHub CLI (if available)
gh pr create --title "feat: implement new feature" --body "Description of changes"

# Or create PR through GitHub web interface
```

## 🔄 Release Process

### Automatic Releases

- **Merges to `main`**: Creates pre-release versions (e.g., `v1.3.0-alpha.1`)
- **Release branches**: Creates stable releases (e.g., `v1.2.0`, `v1.2.1`)

### Manual Release Testing

```bash
# Test what the next release would be
make release-dry-run

# Check current version
make version
```

### Creating a Release Branch

```bash
# When ready for a new major.minor release
git checkout main
git pull origin main
git checkout -b release/v1.3

# Push the release branch
git push origin release/v1.3

# The CI/CD pipeline will automatically create v1.3.0
```

## 🔧 Troubleshooting

### Commit Message Rejected

If your commit message is rejected:

1. **Check the format:**
   ```
   type(scope): description
   ```

2. **Valid types:** feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert

3. **Use interactive commit:**
   ```bash
   make commit
   ```

### Fix a Rejected Commit

```bash
# Amend the last commit message
git commit --amend -m "feat(api): correct commit message format"

# For multiple commits, use interactive rebase
git rebase -i HEAD~3
```

### Hotfix Process

```bash
# Critical bug in production
git checkout main
git checkout -b fix/critical-security-patch

# Make the fix
git commit -m "fix(security): patch authentication vulnerability"

# Create PR to main first
gh pr create --title "fix: critical security patch"

# After merge to main, cherry-pick to release branch
git checkout release/v1.2
git cherry-pick <commit-hash>
git push origin release/v1.2
```

## 📊 Monitoring Releases

### GitHub Releases
- Automatically created with changelogs
- Includes compiled binaries
- Docker images published to `ghcr.io/grammatonic/pihole-network-analyzer`

### Version Information
```bash
# Current version info
make version

# Check if semantic-release would create a release
make release-dry-run
```

## 🔗 References

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [GitLab Flow](https://docs.gitlab.com/ee/topics/gitlab_flow.html)
- [Project Branching Strategy](./BRANCHING_STRATEGY.md)
