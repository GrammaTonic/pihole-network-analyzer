#!/bin/bash

# GitHub Branch Protection Setup Script
# Sets up recommended branch protection rules for Pi-hole Network Analyzer

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_OWNER="GrammaTonic"
REPO_NAME="pihole-network-analyzer"
PROTECTED_BRANCH="main"
FALLBACK_BRANCH="master"

# Status checks that must pass before merging
REQUIRED_STATUS_CHECKS=(
    "test"
    "validate-integration-tests"
    "integration-test"
    "security"
)

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if GitHub CLI is installed
check_gh_cli() {
    if ! command -v gh &> /dev/null; then
        print_error "GitHub CLI (gh) is not installed."
        print_status "Install it from: https://cli.github.com/"
        exit 1
    fi
    
    print_success "GitHub CLI found"
}

# Check if user is authenticated with GitHub CLI
check_gh_auth() {
    if ! gh auth status &> /dev/null; then
        print_error "Not authenticated with GitHub CLI."
        print_status "Run 'gh auth login' to authenticate"
        exit 1
    fi
    
    print_success "GitHub CLI authentication verified"
}

# Determine the default branch
get_default_branch() {
    local default_branch
    default_branch=$(gh api repos/${REPO_OWNER}/${REPO_NAME} --jq '.default_branch' 2>/dev/null || echo "")
    
    if [[ -n "$default_branch" ]]; then
        echo "$default_branch"
    else
        # Fallback: check if main or master exists
        if gh api repos/${REPO_OWNER}/${REPO_NAME}/branches/main &>/dev/null; then
            echo "main"
        elif gh api repos/${REPO_OWNER}/${REPO_NAME}/branches/master &>/dev/null; then
            echo "master"
        else
            print_error "Could not determine default branch"
            exit 1
        fi
    fi
}

# Build required status checks JSON
build_status_checks_json() {
    local contexts=""
    for check in "${REQUIRED_STATUS_CHECKS[@]}"; do
        if [[ -n "$contexts" ]]; then
            contexts="${contexts},"
        fi
        contexts="${contexts}\"${check}\""
    done
    
    echo "{\"strict\":true,\"contexts\":[${contexts}]}"
}

# Apply branch protection rules
apply_protection() {
    local branch="$1"
    local status_checks_json
    status_checks_json=$(build_status_checks_json)
    
    print_status "Applying branch protection to '${branch}' branch..."
    
    # Create the protection rule using GitHub API
    if gh api repos/${REPO_OWNER}/${REPO_NAME}/branches/${branch}/protection \
        --method PUT \
        --field required_status_checks="${status_checks_json}" \
        --field enforce_admins=true \
        --field required_pull_request_reviews='{"required_approving_review_count":1,"dismiss_stale_reviews":true,"require_code_owner_reviews":false}' \
        --field restrictions=null \
        --field required_linear_history=true \
        --field allow_force_pushes=false \
        --field allow_deletions=false &>/dev/null; then
        
        print_success "Branch protection rules applied successfully to '${branch}'"
        return 0
    else
        print_error "Failed to apply branch protection rules to '${branch}'"
        return 1
    fi
}

# Verify protection rules are active
verify_protection() {
    local branch="$1"
    
    print_status "Verifying branch protection rules for '${branch}'..."
    
    local protection_info
    if protection_info=$(gh api repos/${REPO_OWNER}/${REPO_NAME}/branches/${branch}/protection 2>/dev/null); then
        print_success "Branch protection is active for '${branch}'"
        
        # Show key protection settings
        echo
        print_status "Current protection settings:"
        echo "  • Required status checks: $(echo "$protection_info" | jq -r '.required_status_checks.contexts[]' | tr '\n' ', ' | sed 's/,$//')"
        echo "  • Require PR reviews: $(echo "$protection_info" | jq -r '.required_pull_request_reviews.required_approving_review_count') reviewer(s)"
        echo "  • Enforce for admins: $(echo "$protection_info" | jq -r '.enforce_admins.enabled')"
        echo "  • Require linear history: $(echo "$protection_info" | jq -r '.required_linear_history.enabled')"
        echo "  • Allow force pushes: $(echo "$protection_info" | jq -r '.allow_force_pushes.enabled')"
        echo "  • Allow deletions: $(echo "$protection_info" | jq -r '.allow_deletions.enabled')"
        echo
        
        return 0
    else
        print_warning "Could not verify branch protection for '${branch}'"
        return 1
    fi
}

# Main execution
main() {
    echo
    print_status "🔐 Pi-hole Network Analyzer - Branch Protection Setup"
    print_status "=================================================="
    echo
    
    # Pre-flight checks
    check_gh_cli
    check_gh_auth
    
    # Determine the default branch
    print_status "Determining default branch..."
    local default_branch
    default_branch=$(get_default_branch)
    print_success "Default branch: ${default_branch}"
    
    # Apply protection
    echo
    if apply_protection "$default_branch"; then
        echo
        verify_protection "$default_branch"
        
        echo
        print_success "🎉 Branch protection setup completed!"
        print_status "Next steps:"
        echo "  1. Test the protection by creating a pull request"
        echo "  2. Verify that merges are blocked until CI passes"
        echo "  3. Confirm that force pushes to ${default_branch} are prevented"
        echo "  4. Review the settings in GitHub repository settings"
        echo
        print_status "Documentation: .github/BRANCH_PROTECTION.md"
        
    else
        echo
        print_error "❌ Branch protection setup failed!"
        print_status "Manual setup required. See .github/BRANCH_PROTECTION.md for detailed instructions."
        exit 1
    fi
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [--help|--dry-run]"
        echo
        echo "Sets up GitHub branch protection rules for Pi-hole Network Analyzer"
        echo
        echo "Options:"
        echo "  --help, -h    Show this help message"
        echo "  --dry-run     Show what would be done without making changes"
        echo
        echo "Requirements:"
        echo "  - GitHub CLI (gh) installed and authenticated"
        echo "  - Repository admin permissions"
        echo
        echo "For more information, see .github/BRANCH_PROTECTION.md"
        exit 0
        ;;
    --dry-run)
        print_status "🔍 DRY RUN MODE - No changes will be made"
        print_status "Would apply the following protection rules:"
        echo
        echo "Repository: ${REPO_OWNER}/${REPO_NAME}"
        echo "Branch: main or master (auto-detected)"
        echo "Required status checks: ${REQUIRED_STATUS_CHECKS[*]}"
        echo "Required PR reviews: 1"
        echo "Enforce for admins: Yes"
        echo "Require linear history: Yes"
        echo "Allow force pushes: No"
        echo "Allow deletions: No"
        echo
        print_status "Run without --dry-run to apply these settings"
        exit 0
        ;;
    "")
        # No arguments, proceed with normal execution
        main
        ;;
    *)
        print_error "Unknown argument: $1"
        print_status "Use --help for usage information"
        exit 1
        ;;
esac