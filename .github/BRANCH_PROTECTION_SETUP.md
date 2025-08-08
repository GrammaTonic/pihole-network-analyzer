# Branch Protection Setup - Quick Start Guide

This guide provides step-by-step instructions for setting up GitHub branch protection on the Pi-hole Network Analyzer repository.

## 🎯 What This Accomplishes

The branch protection setup provides:

✅ **Security**: Prevents direct pushes to main branch  
✅ **Quality**: Requires all CI tests to pass before merging  
✅ **Collaboration**: Mandates code reviews for all changes  
✅ **Reliability**: Ensures conversation resolution before merging  
✅ **History**: Maintains clean, linear commit history  

## 🚀 Quick Setup (2 minutes)

### Prerequisites
- Repository admin access to `GrammaTonic/pihole-network-analyzer`
- GitHub CLI installed (`gh`) and authenticated

### One-Command Setup
```bash
# Navigate to repository
cd pihole-network-analyzer

# Run the automated setup script
./scripts/setup-branch-protection.sh
```

That's it! The script will:
1. Detect the default branch (main/master)
2. Apply comprehensive protection rules
3. Verify the configuration
4. Display current settings

### Verify Setup
After running the script, you should see:
```
✅ Branch protection is active for 'main'

Current protection settings:
  • Required status checks: test, validate-integration-tests, integration-test, security
  • Require PR reviews: 1 reviewer(s)
  • Enforce for admins: true
  • Require linear history: true
  • Allow force pushes: false
  • Allow deletions: false
```

## 🧪 Testing Branch Protection

### Test 1: Direct Push Prevention
```bash
# This should be blocked
git checkout main
echo "test" >> README.md
git commit -am "test direct push"
git push origin main
# Expected: Push rejected by branch protection
```

### Test 2: PR Workflow
```bash
# This should work
git checkout -b test-branch-protection
echo "testing" >> test.txt
git add test.txt
git commit -m "test: verify branch protection workflow"
git push origin test-branch-protection
# Open PR in GitHub - should require CI and review
```

## 📋 Protection Rules Applied

| Setting | Value | Purpose |
|---------|-------|---------|
| **Required Status Checks** | test, validate-integration-tests, integration-test, security | Ensures code quality and security |
| **Require PR Reviews** | 1 reviewer | Human oversight for all changes |
| **Dismiss Stale Reviews** | Yes | Re-review when new commits added |
| **Require Code Owner Review** | Yes | CODEOWNERS file specifies reviewers |
| **Require Conversation Resolution** | Yes | All feedback must be addressed |
| **Require Linear History** | Yes | Clean, readable commit history |
| **Restrict Force Pushes** | Yes | Prevents history rewriting |
| **Restrict Deletions** | Yes | Prevents accidental branch deletion |
| **Enforce for Admins** | Yes | Rules apply to everyone |

## 🔍 Manual Setup (Alternative)

If you prefer manual setup via GitHub web interface:

1. **Go to Repository Settings**
   - Navigate to `https://github.com/GrammaTonic/pihole-network-analyzer/settings`
   - Click "Branches" in the left sidebar

2. **Add Branch Protection Rule**
   - Click "Add rule"
   - Branch name pattern: `main` (or `master`)

3. **Configure Protection Settings**
   - ☑️ Require a pull request before merging
     - ☑️ Require approvals: 1
     - ☑️ Dismiss stale pull request approvals
     - ☑️ Require review from code owners
   - ☑️ Require status checks to pass before merging
     - ☑️ Require branches to be up to date
     - Add: `test`, `validate-integration-tests`, `integration-test`, `security`
   - ☑️ Require conversation resolution before merging
   - ☑️ Require linear history
   - ☑️ Restrict pushes that create files larger than 100 MB
   - ☑️ Restrict force pushes
   - ☑️ Restrict deletions
   - ☑️ Do not allow bypassing the above settings

4. **Save Protection Rule**
   - Click "Create" to apply the rules

## 🛠️ Maintenance

### Adding New Status Checks
When adding new CI jobs that should block merging:

1. **Update the script**: Edit `scripts/setup-branch-protection.sh`
   ```bash
   REQUIRED_STATUS_CHECKS=(
       "test"
       "validate-integration-tests"
       "integration-test"
       "security"
       "new-check-name"  # Add here
   )
   ```

2. **Re-run setup**: `./scripts/setup-branch-protection.sh`

3. **Update documentation**: Update this guide and `.github/BRANCH_PROTECTION.md`

### Emergency Bypass
For urgent hotfixes:
1. Repository admins can temporarily disable protection
2. Apply hotfix directly to main branch
3. Re-enable protection immediately
4. Document the exception in commit messages

## 📚 Related Documentation

- **[Branch Protection Details](.github/BRANCH_PROTECTION.md)** - Complete configuration reference
- **[Contributing Guide](CONTRIBUTING.md)** - Developer workflow with branch protection
- **[PR Template](.github/PULL_REQUEST_TEMPLATE.md)** - Structured pull request format
- **[Issue Templates](.github/ISSUE_TEMPLATE/)** - Bug reports and feature requests

## 🆘 Troubleshooting

### Script Fails: "GitHub CLI not found"
```bash
# Install GitHub CLI
# macOS: brew install gh
# Ubuntu: sudo apt install gh
# Windows: winget install GitHub.cli

# Authenticate
gh auth login
```

### Script Fails: "Not authenticated"
```bash
gh auth login
# Follow the prompts to authenticate
```

### Script Fails: "API rate limit"
```bash
# Wait a few minutes and retry
./scripts/setup-branch-protection.sh
```

### Protection Not Working
1. Check if rules are active: Go to repository Settings > Branches
2. Verify required status checks match CI job names
3. Confirm you're testing against the protected branch (main/master)

### CI Jobs Not Listed as Required
1. The CI jobs must exist and run at least once
2. Job names are case-sensitive
3. Update the required checks list if CI job names changed

## ✅ Success Indicators

After successful setup, you should see:

- **GitHub Settings**: Branch protection rules visible in repository settings
- **Pull Requests**: "Merge" button disabled until all checks pass and review approved
- **Status Checks**: Required checks shown on PR with pass/fail status
- **Direct Pushes**: Blocked with protection message

## 🎉 Next Steps

1. **Test the workflow** with a sample pull request
2. **Update team documentation** about the new process
3. **Train contributors** on the protected branch workflow
4. **Monitor CI performance** to ensure reasonable check times
5. **Review protection settings** quarterly for optimization

The branch protection is now active and will ensure code quality, security, and collaborative development! 🛡️