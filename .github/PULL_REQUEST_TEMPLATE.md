<!-- 
Thank you for contributing to Pi-hole Network Analyzer! 
Please fill out this template to help us review your changes.
-->

## 🎯 Description

<!-- Provide a clear and concise description of what this PR does -->

## 🔗 Related Issues

<!-- Link to any related issues using keywords: fixes #123, closes #456, relates to #789 -->
- Fixes #
- Closes #
- Related to #

## 🧪 Type of Change

<!-- Mark the relevant option with an [x] -->

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📚 Documentation update
- [ ] 🔧 Internal refactoring (no functional changes)
- [ ] 🧪 Test improvements
- [ ] 🔄 CI/CD improvements

## 🧪 Testing

<!-- Describe the tests you ran and provide instructions to reproduce -->

### Test Commands Run
```bash
# Check all that you've run:
[ ] make ci-test           # CI tests
[ ] make feature-branch    # Full feature validation
[ ] make multi-build       # Cross-platform builds
[ ] ./scripts/integration-test.sh  # Integration tests
[ ] Manual testing with real data
```

### Test Environment
- **OS**: <!-- e.g., Ubuntu 22.04, macOS 13, Windows 11 -->
- **Go Version**: <!-- e.g., 1.24.5 -->
- **Terminal**: <!-- e.g., gnome-terminal, iTerm2, Windows Terminal -->

### Test Results
<!-- Describe what you tested and the results -->

**CSV Analysis Testing:**
- [ ] Tested with large CSV files (>100MB)
- [ ] Verified colorized output works correctly
- [ ] Confirmed --no-color and --quiet modes work
- [ ] Validated report generation

**Pi-hole Integration Testing:**
- [ ] Tested SSH connection functionality
- [ ] Verified database query execution
- [ ] Confirmed hostname resolution works
- [ ] Validated hardware address mapping

## 📋 Checklist

<!-- Mark completed items with [x] -->

### Code Quality
- [ ] My code follows the project's style guidelines (`make fmt`)
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] My changes generate no new warnings (`make vet`)

### Testing
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally (`make test`)
- [ ] I have tested the colorized output functionality
- [ ] I have tested with both `--no-color` and default modes

### Documentation
- [ ] I have made corresponding changes to the documentation
- [ ] I have updated the README.md if needed
- [ ] I have added/updated inline code comments where necessary

### Compatibility
- [ ] My changes work on Linux, macOS, and Windows
- [ ] I have tested cross-platform builds (`make multi-build`)
- [ ] My changes are compatible with both CSV and Pi-hole modes
- [ ] I have considered terminal compatibility (colors, emojis)

## 🖼️ Screenshots/Examples

<!-- If your changes affect the UI/output, include before/after screenshots -->

### Before
<!-- Screenshot or example output before your changes -->

### After
<!-- Screenshot or example output after your changes -->

## 🔄 Branch Protection Requirements

<!-- These will be automatically checked by branch protection rules -->

This PR will be automatically validated by:
- ✅ **CI Status Checks** - All tests must pass
- ✅ **Code Review** - Maintainer approval required
- ✅ **Up-to-date Branch** - Must be current with main
- ✅ **Conversation Resolution** - All discussions must be resolved

See [Branch Protection Configuration](.github/BRANCH_PROTECTION.md) for details.

## 📝 Additional Notes

<!-- Add any additional information that reviewers should know -->

### Performance Impact
<!-- Describe any performance implications of your changes -->

### Breaking Changes
<!-- List any breaking changes and migration steps required -->

### Future Work
<!-- Mention any follow-up work or related improvements -->

---

<!-- 
Automated checks will run when you create this PR:
- Unit and integration tests
- Security vulnerability scanning  
- Cross-platform build verification
- Code formatting and linting

The PR will be ready to merge when:
✅ All automated checks pass
✅ Code review is approved
✅ All conversations are resolved
✅ Branch is up-to-date with main
-->