# Changelog

All notable changes to this project will be documented in this file. See [Conventional Commits](https://conventionalcommits.org) for commit guidelines.

## 1.0.0 (2025-08-12)


### ⚠ BREAKING CHANGES

* Enhanced dashboard introduces new API endpoints and WebSocket integration
* **web:** Web server now includes WebSocket endpoints and requires
gorilla/websocket dependency
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

* Enhanced Dashboard with Chart.js Integration and Real-time WebSocket Updates (#18) ([654ae75](https://github.com/GrammaTonic/pihole-network-analyzer/commit/654ae7583de87de2b6b517ae57079803a1cfdb47)), closes [#18](https://github.com/GrammaTonic/pihole-network-analyzer/issues/18)


### docs

* **release:** implement semantic versioning and release automation framework ([003c765](https://github.com/GrammaTonic/pihole-network-analyzer/commit/003c76516638c7ea46f584db569bbd97a28b4e7a))


### 🚀 Features

* add colorized terminal output support ([e729e63](https://github.com/GrammaTonic/pihole-network-analyzer/commit/e729e63afb48c6e2fc0be5390af9d810e36941f9))
* Add comprehensive AI coding assistant instructions ([a545a03](https://github.com/GrammaTonic/pihole-network-analyzer/commit/a545a03186e6be997e412ce9d75ce2ee6559bd9f))
* Add comprehensive colorized output functionality with integration testing ([d276a8d](https://github.com/GrammaTonic/pihole-network-analyzer/commit/d276a8d769a27409347dfabb7a8f9e4feca585a8))
* Add DNS Usage Analyzer tool with CSV and Pi-hole support ([bb5ee14](https://github.com/GrammaTonic/pihole-network-analyzer/commit/bb5ee148cdb1976b5db162bf3427fe0bdc6fddd6))
* Add structured logging to replace fmt.Printf statements ([#5](https://github.com/GrammaTonic/pihole-network-analyzer/issues/5)) ([720257d](https://github.com/GrammaTonic/pihole-network-analyzer/commit/720257d9432f25a76dd42d91170ba463427d15ed))
* Add test configuration and mock data for DNS Analyzer ([e12b5b8](https://github.com/GrammaTonic/pihole-network-analyzer/commit/e12b5b8a94d5492600c90a06f514418445547bf9))
* Complete ML engine implementation and CLI integration ([55f2fb6](https://github.com/GrammaTonic/pihole-network-analyzer/commit/55f2fb6c1892a5392c46c4ebe11314a49aef43f6))
* Enhance DNS Usage Analyzer with Pi-hole support and usage instructions ([5dc719e](https://github.com/GrammaTonic/pihole-network-analyzer/commit/5dc719e8bd5f5882eb5b87bcbb8de47a647b7bcc))
* Implement comprehensive ML system for anomaly detection and trend analysis ([aecf4b3](https://github.com/GrammaTonic/pihole-network-analyzer/commit/aecf4b31ece651614297f43b0e0267af0ed69bf1))
* integrate Phase 5 enhanced analyzer with API-first architecture ([7fe16b1](https://github.com/GrammaTonic/pihole-network-analyzer/commit/7fe16b14cfb0c6eca439817e32cf31a972ab425b))
* mainstream Docker API-only configuration ([c8ee355](https://github.com/GrammaTonic/pihole-network-analyzer/commit/c8ee3556db0ea6640d1063ae52c78f9be65a75d8))
* Major refactoring and standardization release ([e326a71](https://github.com/GrammaTonic/pihole-network-analyzer/commit/e326a711fa6aba9d94f09be42cc07c496d6f7f15))
* Make Docker the primary installation method with comprehensive environment variable support ([#20](https://github.com/GrammaTonic/pihole-network-analyzer/issues/20)) ([8d7ccb7](https://github.com/GrammaTonic/pihole-network-analyzer/commit/8d7ccb79179f75dd196e118a761dfeb9bb0268ec))
* prepare for v1.0.0 production release ([1bbeb23](https://github.com/GrammaTonic/pihole-network-analyzer/commit/1bbeb235fe6dde73d99434d5667cc286228f4f2e))
* **release:** implement unified Git branching and semantic versioning framework ([a7942c4](https://github.com/GrammaTonic/pihole-network-analyzer/commit/a7942c471888519e54fe388429a206e48f210e2d))
* separate IP and MAC address into distinct table columns ([c06dd12](https://github.com/GrammaTonic/pihole-network-analyzer/commit/c06dd1222b29817dba0aa5c5e65932548eb3c8e3))
* Speed - Fast builds with caching ([#6](https://github.com/GrammaTonic/pihole-network-analyzer/issues/6)) ([54d9fc0](https://github.com/GrammaTonic/pihole-network-analyzer/commit/54d9fc02c78ed50a642921be311b44c59800ebf3))
* Update GitHub Actions workflows to use latest setup-go and cache actions; add custom configuration files for DNS Analyzer ([29daf16](https://github.com/GrammaTonic/pihole-network-analyzer/commit/29daf16f57a8e954287fa9c62e873a265a287b7a))
* **web:** implement comprehensive WebSocket real-time updates system ([#16](https://github.com/GrammaTonic/pihole-network-analyzer/issues/16)) ([5150d71](https://github.com/GrammaTonic/pihole-network-analyzer/commit/5150d71690a232b3bddd221a6522c9286d97b3bc))


### 🐛 Bug Fixes

* CI pipeline failures with colorized integration tests ([2c2addd](https://github.com/GrammaTonic/pihole-network-analyzer/commit/2c2addd90a01285a94595db64280aaf81bb97c35))
* **ci:** Add registry authentication to test-containers job ([2f30de8](https://github.com/GrammaTonic/pihole-network-analyzer/commit/2f30de8f545b723874de1738fd373b59533daafd))
* **ci:** use standard GitHub token for release workflow ([b31a1d8](https://github.com/GrammaTonic/pihole-network-analyzer/commit/b31a1d8dd8ce01ac49a60745c43318932cab0047))
* complete domain colorization and document colorized output ([b377d58](https://github.com/GrammaTonic/pihole-network-analyzer/commit/b377d58b4539d90480b4f2f22c67f19c893037b0))
* correct image reference in security scan ([8529915](https://github.com/GrammaTonic/pihole-network-analyzer/commit/852991528ee6cc6ea9bc618cb4293b5b11948ac3))
* correct table structure alignment ([044ef79](https://github.com/GrammaTonic/pihole-network-analyzer/commit/044ef791a58631318ffcf09118a20f02757aeb31))
* **deps:** include package-lock.json for CI dependency caching ([f1808bc](https://github.com/GrammaTonic/pihole-network-analyzer/commit/f1808bc8453ae4ceb00508e65790a654d1d7bd69))
* make security scan upload non-blocking ([0457e87](https://github.com/GrammaTonic/pihole-network-analyzer/commit/0457e870f2707a30b62bdbfa629ade4bb4ef17f4))
* **release:** update branch pattern for v1.0 release branch ([cb7612f](https://github.com/GrammaTonic/pihole-network-analyzer/commit/cb7612f943c1d419270f24e00b500619ba19d543))
* remove duplicate registry prefix in security scan ([9f80a2e](https://github.com/GrammaTonic/pihole-network-analyzer/commit/9f80a2e6c19d49b7c6a3dd60b80bb277f289983f))
* resolve code formatting issues for CI pipeline ([5edb4f8](https://github.com/GrammaTonic/pihole-network-analyzer/commit/5edb4f82fb1593afc4ec8fe62adbab1ba3252985))
* resolve colorized output test failures in CI pipeline ([7665d70](https://github.com/GrammaTonic/pihole-network-analyzer/commit/7665d700f4b3e2b28fe837abbd8d92d857d2f23d))
* resolve CSV Analysis integration test failures ([118aa9d](https://github.com/GrammaTonic/pihole-network-analyzer/commit/118aa9d2e26a3f834e92bc03c8b0317f13317efa))
* resolve Go version compatibility and format string issues ([a02b478](https://github.com/GrammaTonic/pihole-network-analyzer/commit/a02b478b17801d69ae22362839dac559831e22e6))
* resolve integration test failures in CI pipeline ([7f440ee](https://github.com/GrammaTonic/pihole-network-analyzer/commit/7f440eec9846cde3c7fa49d47ea2df0c06fd0b23))
* resolve Pi-hole DB Analysis pipeline failure ([a950106](https://github.com/GrammaTonic/pihole-network-analyzer/commit/a9501068936d6e0b85cf53ba0810612ed184ae8b))
* update CI pipeline Go versions for compatibility ([7fee279](https://github.com/GrammaTonic/pihole-network-analyzer/commit/7fee2795c6e1b6ae3398b2be8dbadb1095d6c40a))


### ♻️ Code Refactoring

* remove legacy phase references from codebase ([5ab9fb9](https://github.com/GrammaTonic/pihole-network-analyzer/commit/5ab9fb9eba6646d9d471b9bfcb689f1def234855))
* Remove obsolete workflow files and streamline CI/CD configurations ([9ce2286](https://github.com/GrammaTonic/pihole-network-analyzer/commit/9ce22869d239a6b024b5b50aabb62405a5dfbd89))


### 🔒 Security

* fix Go standard library vulnerability GO-2025-3849 ([1c3ad3e](https://github.com/GrammaTonic/pihole-network-analyzer/commit/1c3ad3eddb4611e9f7e0f71801b30b8bdb3ce3b5))

## [2.1.0](https://github.com/GrammaTonic/pihole-network-analyzer/compare/v2.0.0...v2.1.0) (2025-08-11)


### 🚀 Features

* Make Docker the primary installation method with comprehensive environment variable support ([#20](https://github.com/GrammaTonic/pihole-network-analyzer/issues/20)) ([4f1952a](https://github.com/GrammaTonic/pihole-network-analyzer/commit/4f1952a74709162ab8d7e2315c4d22c41ccb5f9a))

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
