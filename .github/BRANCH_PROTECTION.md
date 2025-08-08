# GitHub Branch Protection Configuration

This document outlines the recommended branch protection settings for the Pi-hole Network Analyzer repository to maintain code quality and ensure a safe development workflow.

## Overview

Branch protection rules help enforce a disciplined development workflow by:
- Requiring pull request reviews before merging
- Ensuring all CI status checks pass before merging
- Preventing accidental force pushes or deletions
- Maintaining a clean commit history

## Recommended Branch Protection Settings

### Main/Master Branch Protection

The following settings should be applied to the `main` (or `master`) branch:

#### Required Status Checks
All status checks must pass before merging:
- `test` - Unit and integration tests
- `validate-integration-tests` - Integration test framework validation  
- `integration-test` - Comprehensive integration test matrix
- `security` - Security vulnerability scanning
- `build-check` - Build verification (feature branches only)

#### Pull Request Requirements
- ✅ **Require pull request reviews before merging**
  - Required number of reviewers: **1**
  - Dismiss stale reviews when new commits are pushed: **Yes**
  - Require review from code owners: **Yes** (when CODEOWNERS file is present)

- ✅ **Require status checks to pass before merging** 
  - Require branches to be up to date before merging: **Yes**
  - Required status checks: All CI jobs listed above

- ✅ **Require conversation resolution before merging**
  - All conversations must be resolved before merge

#### Additional Restrictions
- ✅ **Restrict pushes that create files larger than 100MB**
- ✅ **Restrict force pushes** - Prevent force pushes to protected branch
- ✅ **Restrict deletions** - Prevent deletion of protected branch
- ✅ **Require linear history** - Ensure clean commit history
- ✅ **Do not allow bypassing the above settings** - Apply to administrators too

#### Branch Management
- ✅ **Allow auto-merge** - Enable auto-merge when all requirements are met
- ✅ **Allow squash merging** - Preferred merge method for clean history
- ✅ **Allow merge commits** - Allow traditional merge commits when appropriate
- ❌ **Allow rebase merging** - Disabled to prevent complex merge conflicts

## Quick Setup Script

A GitHub CLI script is provided for automated setup:

```bash
# Run the branch protection setup script
./scripts/setup-branch-protection.sh

# Or manually using GitHub CLI
gh api repos/:owner/:repo/branches/main/protection \
  --method PUT \
  --field required_status_checks='{"strict":true,"contexts":["test","validate-integration-tests","integration-test","security"]}' \
  --field enforce_admins=true \
  --field required_pull_request_reviews='{"required_approving_review_count":1,"dismiss_stale_reviews":true}' \
  --field restrictions=null
```

## Rationale

### Why These Settings?

1. **Required Status Checks**: Ensures all CI tests pass, maintaining code quality and preventing regressions
2. **Pull Request Reviews**: Human oversight catches issues automated tests might miss
3. **Conversation Resolution**: Ensures all feedback is addressed before merging
4. **Force Push Protection**: Prevents accidental history rewriting that could cause data loss
5. **Linear History**: Maintains a clean, readable commit history for easier debugging

### Compatibility with Current Workflow

These settings work seamlessly with the existing CI/CD pipeline:

- **Feature Branches**: Run all tests and build checks before allowing merge
- **Main Branch**: Triggers full builds and releases after successful merge
- **Integration Tests**: Matrix testing across multiple scenarios and Go versions
- **Security Scanning**: Automated vulnerability detection before merge

## Testing Branch Protection

After applying these settings:

1. **Create a test branch** and make a small change
2. **Open a pull request** to verify required checks appear
3. **Confirm merge is blocked** until all checks pass and reviews are provided
4. **Test that force pushes are prevented** on the main branch
5. **Verify auto-merge works** when all requirements are met

## Maintenance

### Regular Review
- Review branch protection settings quarterly
- Update required status checks when CI jobs change
- Adjust reviewer requirements based on team size

### Adding New Status Checks
When adding new CI jobs that should block merging:

1. Add the job name to required status checks list
2. Update this documentation
3. Test with a sample pull request

### Emergency Procedures
In case of urgent hotfixes:
- Repository administrators can temporarily disable protection
- Re-enable protection immediately after emergency merge
- Document the exception in commit message

## Related Documentation

- [Contribution Guidelines](CONTRIBUTING.md) - Developer workflow and standards
- [Feature Workflow](docs/FEATURE_WORKFLOW.md) - Detailed development process
- [CI/CD Pipeline](.github/workflows/ci.yml) - Automated testing and building
- [Integration Testing](docs/INTEGRATION_TESTING_GUIDE.md) - Test framework details

## Support

For questions about branch protection or to request changes:
- Open an issue with the `documentation` label
- Contact repository maintainers
- Review GitHub's [branch protection documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/defining-the-mergeability-of-pull-requests/about-protected-branches)