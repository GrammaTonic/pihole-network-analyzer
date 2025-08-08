# Contributing to Pi-hole Network Analyzer

Thank you for your interest in contributing to the Pi-hole Network Analyzer! This document provides guidelines for contributing to ensure a smooth development process.

## 🛡️ Branch Protection & Development Workflow

This repository uses **branch protection rules** to maintain code quality and ensure safe collaboration:

- **All changes must go through pull requests** - direct pushes to main branch are blocked
- **CI status checks must pass** before merging (tests, builds, security scans)
- **Pull request reviews are required** before merging
- **Conversation resolution is mandatory** before merging
- **Force pushes and branch deletions are prevented** on main branch

For detailed information, see [Branch Protection Configuration](.github/BRANCH_PROTECTION.md).

## 🚀 Development Process

### 1. Fork and Clone
```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/pihole-network-analyzer.git
cd pihole-network-analyzer

# Add upstream remote
git remote add upstream https://github.com/GrammaTonic/pihole-network-analyzer.git
```

### 2. Create a Feature Branch
```bash
# Create and switch to a new feature branch
git checkout -b feature/your-feature-name

# Keep your branch up to date with upstream
git fetch upstream
git rebase upstream/main
```

### 3. Development Standards

#### Code Quality
- **Go Formatting**: Run `make fmt` before committing
- **Linting**: Run `make vet` to check for issues  
- **Testing**: Run `make ci-test` to validate your changes
- **Build Verification**: Run `make feature-branch` for comprehensive validation

#### Testing Requirements
- **Unit Tests**: Add tests for new functionality in `tests/unit/`
- **Integration Tests**: Update integration tests if needed in `tests/integration/`
- **CI Compatibility**: Ensure tests pass in CI environment

#### Documentation
- **Code Comments**: Add meaningful comments for complex logic
- **README Updates**: Update documentation for new features
- **API Changes**: Document any breaking changes

### 4. Commit Guidelines

#### Commit Message Format
Use conventional commit format:
```
type(scope): description

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `ci`: CI/CD changes

**Examples:**
```bash
git commit -m "feat(colors): add new domain highlighting for development sites"
git commit -m "fix(ssh): handle connection timeout gracefully"
git commit -m "docs(readme): update installation instructions"
```

### 5. Pre-Commit Validation

Before pushing your changes:

```bash
# Run comprehensive validation
make feature-branch

# Or run individual checks
make ci-test          # Same tests as CI
make fmt              # Format code
make vet              # Static analysis
make multi-build      # Test cross-platform builds
```

### 6. Submit Pull Request

#### Creating the PR
1. **Push your feature branch** to your fork
2. **Open a pull request** against the main branch
3. **Fill out the PR template** completely
4. **Link related issues** using keywords (fixes #123)

#### PR Requirements
- ✅ **Clear title and description** explaining the changes
- ✅ **All CI checks passing** (automated via branch protection)
- ✅ **Code review approval** from maintainers
- ✅ **Conversations resolved** before merging
- ✅ **Up-to-date with main branch** (will be enforced)

#### PR Review Process
- **Automated checks** run automatically (tests, builds, security)
- **Maintainer review** provides feedback and approval
- **Address feedback** by pushing new commits to your branch
- **Auto-merge** when all requirements are satisfied

## 🎯 Contribution Areas

### High-Priority Areas
- **Refactoring main.go** - Breaking down the monolithic 1693-line file
- **Performance optimization** - Improving large file processing
- **Error handling** - Better SSH connection reliability
- **Testing coverage** - Adding comprehensive test suites

### Feature Opportunities
- **Docker support** - Containerization and orchestration
- **Prometheus metrics** - Real-time monitoring integration
- **API endpoints** - REST API for external integration
- **Configuration management** - Enhanced config validation

### Documentation Needs
- **API documentation** - Code structure and interfaces
- **Deployment guides** - Production deployment scenarios
- **Troubleshooting** - Common issues and solutions
- **Performance tuning** - Optimization recommendations

## 🔧 Development Environment

### Prerequisites
- **Go 1.24+** - Latest Go version for development
- **Make** - Build automation tool
- **Git** - Version control
- **Terminal with color support** - For testing colorized output

### Local Setup
```bash
# Install dependencies
make install-deps

# Build the application
make build

# Run with test data
make run

# Test Pi-hole connection setup
make setup-pihole
```

### IDE Configuration
Recommended settings for Go development:
- **gofmt** on save
- **goimports** for import management
- **go vet** integration
- **Test coverage** highlighting

## 🐛 Bug Reports

### Before Reporting
1. **Search existing issues** to avoid duplicates
2. **Test with latest version** to ensure bug still exists
3. **Gather system information** (OS, Go version, terminal type)

### Bug Report Template
```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Run command '...'
2. With file '...'
3. See error

**Expected behavior**
What you expected to happen.

**Environment:**
- OS: [e.g., Ubuntu 22.04]
- Go version: [e.g., 1.24.5]
- Terminal: [e.g., gnome-terminal, iTerm2]
- Colorized output: [Yes/No]

**Additional context**
Add any other context about the problem here.
```

## 🚀 Feature Requests

### Before Requesting
1. **Check existing issues** and roadmap
2. **Consider the scope** - does it fit the project goals?
3. **Think about implementation** - how would it work?

### Feature Request Template
```markdown
**Is your feature request related to a problem?**
A clear description of what the problem is.

**Describe the solution you'd like**
A clear description of what you want to happen.

**Describe alternatives you've considered**
A clear description of any alternative solutions.

**Additional context**
Add any other context or screenshots about the feature request.
```

## 📝 Documentation Contributions

### Areas for Improvement
- **Installation guides** for different platforms
- **Configuration examples** for various Pi-hole setups
- **Performance optimization** recommendations
- **Troubleshooting guides** for common issues

### Documentation Standards
- **Clear examples** with expected outputs
- **Cross-platform considerations** (Windows, macOS, Linux)
- **Version compatibility** information
- **Screenshots or terminal output** where helpful

## 🤝 Community Guidelines

### Code of Conduct
- **Be respectful** and inclusive in all interactions
- **Provide constructive feedback** during code reviews
- **Help newcomers** understand the codebase and process
- **Focus on the technical merits** of contributions

### Communication
- **GitHub Issues** for bug reports and feature requests
- **Pull Request discussions** for code-specific feedback
- **Clear, concise communication** in all interactions

## 📚 Additional Resources

- **[Feature Workflow Guide](docs/FEATURE_WORKFLOW.md)** - Detailed development process
- **[Branch Protection](..github/BRANCH_PROTECTION.md)** - Repository protection settings
- **[Integration Testing](docs/INTEGRATION_TESTING_GUIDE.md)** - Test framework documentation
- **[Project Roadmap](docs/ROADMAP_DOCKER_MONITORING.md)** - Future development plans

## 🆘 Getting Help

If you need help:
1. **Check existing documentation** in the `docs/` directory
2. **Search closed issues** for similar problems
3. **Open a new issue** with the `question` label
4. **Be specific** about what you're trying to achieve

Thank you for contributing to the Pi-hole Network Analyzer! 🎉