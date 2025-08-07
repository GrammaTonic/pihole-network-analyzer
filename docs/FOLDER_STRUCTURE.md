# Pi-hole Network Analyzer - Folder Structure

## 📁 Project Organization

```
pihole-network-analyzer/
├── docs/                           # 📚 Documentation
│   ├── README.md                   # Main project documentation
│   ├── FEATURE_WORKFLOW.md         # Development workflow guide
│   ├── INTEGRATION_TESTING.md      # Testing framework documentation
│   ├── INTEGRATION_TESTING_GUIDE.md # Testing usage guide
│   ├── PROJECT_SUMMARY.md          # Project overview
│   ├── TEST_SUMMARY.md            # Test results and coverage
│   ├── IMPLEMENTATION_SUMMARY.md   # Technical implementation details
│   ├── ROADMAP_DOCKER_MONITORING.md # Future Docker/monitoring plans
│   ├── TODO_DOCKER_PROMETHEUS_GRAFANA.md # Docker integration TODO
│   └── MILESTONE_1_DOCKER_QUICKSTART.md # Docker milestone planning
│
├── scripts/                        # 🛠️ Shell Scripts & Automation
│   ├── integration-test.sh         # Integration testing framework
│   ├── ci-test.sh                  # CI/CD testing automation
│   ├── validate-ci.sh              # Local CI validation
│   ├── pre-push-test.sh           # Pre-push validation
│   ├── analyze.sh                  # Analysis helper script
│   ├── test.sh                     # Basic test runner
│   └── usage.sh                    # Usage examples and help
│
├── reports/                        # 📊 Generated Reports (gitignored)
│   └── dns_usage_report_*.txt      # Auto-generated analysis reports
│
├── .github/                        # ⚙️ GitHub Actions & CI/CD
│   └── workflows/
│       └── ci.yml                  # Comprehensive CI/CD pipeline
│
├── .vscode/                        # 🔧 VS Code Configuration
│   └── ...                         # Editor settings and configurations
│
├── main.go                         # 🚀 Main Application Entry Point
├── colors.go                       # 🎨 Colorized Output System
├── config.go                       # ⚙️ Configuration Management
├── test_runner.go                  # 🧪 Test Suite Runner
├── test_data.go                    # 📝 Test Data Generation
├── colors_test.go                  # 🎨 Color System Tests
├── colors_integration_test.go      # 🎨 Color Integration Tests
├── test.csv                        # 📄 Sample DNS Data for Testing
├── custom-config.json              # ⚙️ Custom Configuration Example
├── go.mod & go.sum                 # 📦 Go Module Dependencies
├── Makefile                        # 🔨 Build & Task Automation
└── README.md → docs/README.md      # 📖 Main Documentation (moved)
```

## 🎯 Benefits of This Structure

### 📚 **Documentation (`docs/`)**
- **Centralized**: All documentation in one place
- **Discoverable**: Easy to find and maintain
- **Organized**: Logical grouping by purpose

### 🛠️ **Scripts (`scripts/`)**
- **Automation**: All shell scripts organized together
- **Maintainability**: Easy to update paths and references
- **CI/CD**: Clear separation of automation tools

### 📊 **Reports (`reports/`)**
- **Clean Repository**: Generated files don't clutter the root
- **Gitignored**: Automatic cleanup of temporary files
- **Organized**: All reports in a dedicated location

## 🔄 Migration Impact

### ✅ **Updated References**
- GitHub Actions workflows updated to use `scripts/` paths
- Makefile updated for new script locations
- Documentation cross-references updated
- Internal script references updated

### 🛡️ **Backward Compatibility**
- Core Go files remain in root for build compatibility
- Module structure unchanged (`go.mod` unaffected)
- Binary name and functionality unchanged

### 🚀 **Developer Experience**
- Cleaner root directory
- Logical file organization
- Easier navigation and maintenance
- Better separation of concerns

## 📋 Quick Reference

| Old Location | New Location | Purpose |
|-------------|--------------|---------|
| `*.md` | `docs/*.md` | Documentation files |
| `*.sh` | `scripts/*.sh` | Shell scripts |
| Generated reports | `reports/` | Auto-generated files |
| Core `.go` files | Root | Main application code |

This structure maintains compatibility while providing a cleaner, more professional project organization.
