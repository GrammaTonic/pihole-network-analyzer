# Changelog

All notable changes to this project will be documented in this file. See [Conventional Commits](https://conventionalcommits.org) for commit guidelines.

## [2.0.0](https://github.com/GrammaTonic/pihole-network-analyzer/compare/v1.0.0...v2.0.0) (2025-08-10)


### ⚠ BREAKING CHANGES

* Enhanced dashboard introduces new API endpoints and WebSocket integration
* **web:** Web server now includes WebSocket endpoints and requires
gorilla/websocket dependency

* Enhanced Dashboard with Chart.js Integration and Real-time WebSocket Updates (#18) ([daca396](https://github.com/GrammaTonic/pihole-network-analyzer/commit/daca396eb232de8de1d39212408bb3a60f363c51)), closes [#18](https://github.com/GrammaTonic/pihole-network-analyzer/issues/18)


### 🚀 Features

* **web:** implement comprehensive WebSocket real-time updates system ([#16](https://github.com/GrammaTonic/pihole-network-analyzer/issues/16)) ([61d8745](https://github.com/GrammaTonic/pihole-network-analyzer/commit/61d8745bbfd4a2a2825a8b6132249d0e3bf60980))

## 1.0.0 (2025-08-10)


### ⚠ BREAKING CHANGES

* **release:** Implement comprehensive release pipeline with semantic versioning.

This introduces automated release management with:
- GitLab Flow with Release Branches strategy
- semantic-release automation for version bumping and changelog generation  
- Enhanced GitHub token integration with fallback pattern
- Multi-architecture Docker builds and container registry publishing
- Comprehensive CI/CD pipeline with 17 validation checks
- Repository secrets configuration automation
- Branch protection with proper release branch patterns

This change establishes the foundation for automated v1.0.0 release.

* feat(release): add automated branch protection via GitHub CLI

- Create script to protect release branches using gh CLI
- Add make target: make protect-release-branch VERSION=vX.Y
- Enable protection for main and release/v1.0 branches
- Require PR reviews, CI checks, and restrict force pushes
- Complete branch protection implementation for v1.0.0

* feat(release): complete repository secrets configuration

- Add interactive secrets configuration script
- Create comprehensive secrets management tools
- Update workflow to use enhanced GitHub token
- Add documentation for secrets configuration
- Enable automated publishing capabilities

Features:
- scripts/configure-secrets.sh for guided setup
- make configure-secrets and make secrets-status targets
- Enhanced token fallback in GitHub Actions workflow
- Docker Hub and Slack integration support
- Complete secrets configuration documentation
* This marks the stable v1.0.0 API and feature set. Production ready
release with complete semantic versioning framework, GitLab Flow branching strategy,
and automated CI/CD pipeline.
* **release:** Development workflow now requires conventional commit format.
See docs/QUICK_START_WORKFLOW.md for migration guide.

### docs

* **release:** implement semantic versioning and release automation framework ([9d83a99](https://github.com/GrammaTonic/pihole-network-analyzer/commit/9d83a998653c1ed90d39991dd312eb57b0ac3cb1))


### 🚀 Features

* add colorized terminal output support ([2a6b229](https://github.com/GrammaTonic/pihole-network-analyzer/commit/2a6b229dca2647eeb7a19e3c6e7e5c66a3aed260))
* Add comprehensive AI coding assistant instructions ([37e956a](https://github.com/GrammaTonic/pihole-network-analyzer/commit/37e956a3180f7d491b41b73aa21978609539231d))
* Add comprehensive colorized output functionality with integration testing ([591a37e](https://github.com/GrammaTonic/pihole-network-analyzer/commit/591a37e3c9f4773dd8278ba2dba8b5696fc094bb))
* Add DNS Usage Analyzer tool with CSV and Pi-hole support ([4aa099c](https://github.com/GrammaTonic/pihole-network-analyzer/commit/4aa099ca49276f51b57818057101692b588c9d62))
* Add structured logging to replace fmt.Printf statements ([#5](https://github.com/GrammaTonic/pihole-network-analyzer/issues/5)) ([e8ae148](https://github.com/GrammaTonic/pihole-network-analyzer/commit/e8ae1481241b9db558878a68fc276472bd55e497))
* Add test configuration and mock data for DNS Analyzer ([c1b4343](https://github.com/GrammaTonic/pihole-network-analyzer/commit/c1b4343570775781ce1b833e6f916e84f528e2fe))
* Complete ML engine implementation and CLI integration ([0b6c8aa](https://github.com/GrammaTonic/pihole-network-analyzer/commit/0b6c8aa4c72ea02501d7e1626e284158938aaa2c))
* Enhance DNS Usage Analyzer with Pi-hole support and usage instructions ([04e9282](https://github.com/GrammaTonic/pihole-network-analyzer/commit/04e928282853fc56682b49506339510eb30d83db))
* Implement comprehensive ML system for anomaly detection and trend analysis ([0754e81](https://github.com/GrammaTonic/pihole-network-analyzer/commit/0754e8132df3683fd8d6b8248dbe4ece7b215996))
* integrate Phase 5 enhanced analyzer with API-first architecture ([bf789ea](https://github.com/GrammaTonic/pihole-network-analyzer/commit/bf789ea94cbf70b0b453fdc74044f8943b39b620))
* mainstream Docker API-only configuration ([a040220](https://github.com/GrammaTonic/pihole-network-analyzer/commit/a040220d718222f5bae52a024b0e2ec957e84114))
* Major refactoring and standardization release ([6a4397d](https://github.com/GrammaTonic/pihole-network-analyzer/commit/6a4397decb86d06d0b67b1aef4c0df7a06671974))
* prepare for v1.0.0 production release ([0f55a99](https://github.com/GrammaTonic/pihole-network-analyzer/commit/0f55a99f701b8644c98830bb076b0886d1fad04c))
* **release:** implement unified Git branching and semantic versioning framework ([f728d11](https://github.com/GrammaTonic/pihole-network-analyzer/commit/f728d11f3898b6870a9a2421d97dba078c974ce7))
* separate IP and MAC address into distinct table columns ([81e7aeb](https://github.com/GrammaTonic/pihole-network-analyzer/commit/81e7aeb5a3e5f844d86a63635e427c64bc280652))
* Speed - Fast builds with caching ([#6](https://github.com/GrammaTonic/pihole-network-analyzer/issues/6)) ([1a6003a](https://github.com/GrammaTonic/pihole-network-analyzer/commit/1a6003a8476dc05cd209d6fa2053e75514dc7edd))
* Update GitHub Actions workflows to use latest setup-go and cache actions; add custom configuration files for DNS Analyzer ([24e77f9](https://github.com/GrammaTonic/pihole-network-analyzer/commit/24e77f90b24c0e9a8ced607bd867c335ab4d39e8))


### 🐛 Bug Fixes

* CI pipeline failures with colorized integration tests ([48df7a6](https://github.com/GrammaTonic/pihole-network-analyzer/commit/48df7a6c53b03cf98cd5f7a77c762fc9160c37ed))
* **ci:** Add registry authentication to test-containers job ([75c4187](https://github.com/GrammaTonic/pihole-network-analyzer/commit/75c418791805c01b21624e9091fdf21ad054d0a8))
* **ci:** use standard GitHub token for release workflow ([366799f](https://github.com/GrammaTonic/pihole-network-analyzer/commit/366799f336fa9bff6017620d140c99b1e0da60a8))
* complete domain colorization and document colorized output ([e1e15e7](https://github.com/GrammaTonic/pihole-network-analyzer/commit/e1e15e7d5aabad80976ce200787fe9889cefd48c))
* correct image reference in security scan ([c04e6d1](https://github.com/GrammaTonic/pihole-network-analyzer/commit/c04e6d12f9ac4647fd22577f71099d0a43ad84b1))
* correct table structure alignment ([d99907f](https://github.com/GrammaTonic/pihole-network-analyzer/commit/d99907f62eba7cab7d92dfddc4a233d636057199))
* **deps:** include package-lock.json for CI dependency caching ([ca6fbd7](https://github.com/GrammaTonic/pihole-network-analyzer/commit/ca6fbd7eb3389eefd093348d066709c8577a626a))
* make security scan upload non-blocking ([f4806db](https://github.com/GrammaTonic/pihole-network-analyzer/commit/f4806db650a9dca9c405b1caacc1809935987043))
* **release:** update branch pattern for v1.0 release branch ([9d14f7e](https://github.com/GrammaTonic/pihole-network-analyzer/commit/9d14f7e10a86abc37a748ccfb5a6e95f40927e43))
* remove duplicate registry prefix in security scan ([c59c3d9](https://github.com/GrammaTonic/pihole-network-analyzer/commit/c59c3d958aa5ddf86d961e535a08d693236cb39b))
* resolve code formatting issues for CI pipeline ([d99b294](https://github.com/GrammaTonic/pihole-network-analyzer/commit/d99b2948339ad77b64146bde82a4b5329acf13fa))
* resolve colorized output test failures in CI pipeline ([15810b6](https://github.com/GrammaTonic/pihole-network-analyzer/commit/15810b6ca7b04c4ca396d40e47c21f8d1a6820ef))
* resolve CSV Analysis integration test failures ([81de329](https://github.com/GrammaTonic/pihole-network-analyzer/commit/81de329b69bd04a57bf44f16c1af595a37bcdf65))
* resolve Go version compatibility and format string issues ([19143e6](https://github.com/GrammaTonic/pihole-network-analyzer/commit/19143e6449abc3bf548de9bf04315edc62788811))
* resolve integration test failures in CI pipeline ([54bae11](https://github.com/GrammaTonic/pihole-network-analyzer/commit/54bae110e48d8d9a8020861d62e8facf4e0507a0))
* resolve Pi-hole DB Analysis pipeline failure ([af107e6](https://github.com/GrammaTonic/pihole-network-analyzer/commit/af107e67ba6b5bfceb8aabc1c0745ac5514cbb48))
* update CI pipeline Go versions for compatibility ([ffccc88](https://github.com/GrammaTonic/pihole-network-analyzer/commit/ffccc88099662a4ba78bf1438c3d312e237aeddb))


### ♻️ Code Refactoring

* remove legacy phase references from codebase ([1eef741](https://github.com/GrammaTonic/pihole-network-analyzer/commit/1eef74113f234b121dc1b2dee1d7c3fadc5f3c89))
* Remove obsolete workflow files and streamline CI/CD configurations ([618bf84](https://github.com/GrammaTonic/pihole-network-analyzer/commit/618bf84db605781d9af1d2c931539f7c41a99605))


### 🔒 Security

* fix Go standard library vulnerability GO-2025-3849 ([ed238ae](https://github.com/GrammaTonic/pihole-network-analyzer/commit/ed238ae91ef07912de5371958293c17938a831a4))

## [Unreleased]

### 🚀 Features
- Implemented unified Git branching strategy and semantic versioning framework
- Added automated release pipeline with semantic-release
- Configured conventional commit validation and automation

### 📚 Documentation
- Added comprehensive branching strategy guide
- Created release management documentation
- Updated development workflow documentation

### 🔧 Development Experience
- Added commit message validation with commitlint
- Configured Git hooks for pre-commit checks
- Integrated semantic versioning with CI/CD pipeline

---

*This changelog is automatically generated from conventional commit messages.*
