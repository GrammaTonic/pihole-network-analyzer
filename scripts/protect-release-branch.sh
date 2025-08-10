#!/bin/bash
# Script to enable branch protection for new release branches
# Usage: ./scripts/protect-release-branch.sh v1.1

set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.1"
    exit 1
fi

VERSION=$1
BRANCH_NAME="release/$VERSION"

echo "🔒 Enabling branch protection for $BRANCH_NAME..."

# Check if branch exists
if ! gh api repos/GrammaTonic/pihole-network-analyzer/branches/$BRANCH_NAME > /dev/null 2>&1; then
    echo "❌ Branch $BRANCH_NAME does not exist. Create it first with:"
    echo "   git checkout -b $BRANCH_NAME && git push -u origin $BRANCH_NAME"
    exit 1
fi

# Enable branch protection
gh api repos/GrammaTonic/pihole-network-analyzer/branches/$BRANCH_NAME/protection \
  --method PUT \
  --input - <<EOF
{
  "required_status_checks": {
    "strict": true,
    "checks": [
      {"context": "CI/CD Pipeline"}
    ]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "required_approving_review_count": 1,
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false
  },
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false
}
EOF

echo "✅ Branch protection enabled for $BRANCH_NAME"
echo "📋 Protection settings:"
echo "   - Requires 1 PR review"
echo "   - Requires CI/CD Pipeline status check"
echo "   - Dismisses stale reviews"
echo "   - Enforces rules for admins"
echo "   - Blocks force pushes and deletions"
