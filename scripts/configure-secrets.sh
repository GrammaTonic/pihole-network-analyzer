#!/bin/bash
# Script to configure repository secrets for automated publishing
# Usage: ./scripts/configure-secrets.sh

set -e

echo "🔐 Repository Secrets Configuration for Automated Publishing"
echo "============================================================"
echo ""

# Check if gh CLI is available and authenticated
if ! command -v gh >/dev/null 2>&1; then
    echo "❌ GitHub CLI (gh) not found. Please install it first:"
    echo "   https://cli.github.com/"
    exit 1
fi

echo "✅ GitHub CLI found"

# Check authentication
if ! gh auth status >/dev/null 2>&1; then
    echo "❌ Not authenticated with GitHub. Please run:"
    echo "   gh auth login"
    exit 1
fi

echo "✅ GitHub CLI authenticated"
echo ""

# Check current secrets
echo "📋 Current repository secrets:"
CURRENT_SECRETS=$(gh secret list --json name --jq '.[].name' 2>/dev/null || echo "")

if [ -z "$CURRENT_SECRETS" ]; then
    echo "   No secrets currently configured"
else
    echo "$CURRENT_SECRETS" | sed 's/^/   - /'
fi
echo ""

echo "🔍 Analyzing workflow requirements..."
echo ""

# The workflow uses GITHUB_TOKEN which is automatically provided
echo "✅ GITHUB_TOKEN: Automatically provided by GitHub Actions"
echo "   - Permissions: contents:write, issues:write, pull-requests:write, id-token:write"
echo "   - Used for: GitHub releases, GitHub Container Registry (ghcr.io)"
echo ""

# Check if we need additional secrets
echo "🤔 Optional Enhanced Secrets:"
echo ""

echo "1. ENHANCED_GITHUB_TOKEN (optional):"
echo "   - Enhanced Personal Access Token with fine-grained permissions"
echo "   - Useful for: Better rate limiting, more granular control"
echo "   - Required scopes: repo, write:packages, read:packages"
echo ""

read -p "Do you want to set up an enhanced GitHub token? (y/N): " setup_enhanced

if [[ $setup_enhanced =~ ^[Yy]$ ]]; then
    echo ""
    echo "📋 To create an enhanced Personal Access Token:"
    echo "1. Go to: https://github.com/settings/tokens/new"
    echo "2. Set description: 'Pi-hole Analyzer Release Automation'"
    echo "3. Set expiration: Choose appropriate duration (90 days, 1 year, etc.)"
    echo "4. Select scopes:"
    echo "   ✅ repo (Full control of private repositories)"
    echo "   ✅ write:packages (Upload packages to GitHub Package Registry)"
    echo "   ✅ read:packages (Download packages from GitHub Package Registry)"
    echo "5. Click 'Generate token'"
    echo "6. Copy the token (starts with ghp_)"
    echo ""
    
    read -p "Enter your Personal Access Token (or press Enter to skip): " -s pat_token
    echo ""
    
    if [ -n "$pat_token" ]; then
        echo "🔐 Setting ENHANCED_GITHUB_TOKEN secret..."
        echo "$pat_token" | gh secret set ENHANCED_GITHUB_TOKEN
        echo "✅ ENHANCED_GITHUB_TOKEN secret configured"
    else
        echo "⏭️  Skipped enhanced token setup"
    fi
fi

echo ""
echo "2. DOCKER_REGISTRY_TOKEN (optional):"
echo "   - For publishing to external Docker registries (Docker Hub, etc.)"
echo "   - Current setup uses GitHub Container Registry (ghcr.io)"
echo ""

read -p "Do you want to set up Docker Hub publishing? (y/N): " setup_docker

if [[ $setup_docker =~ ^[Yy]$ ]]; then
    echo ""
    read -p "Enter Docker Hub username: " docker_username
    read -p "Enter Docker Hub access token: " -s docker_token
    echo ""
    
    if [ -n "$docker_username" ] && [ -n "$docker_token" ]; then
        echo "🔐 Setting Docker Hub secrets..."
        echo "$docker_username" | gh secret set DOCKER_USERNAME
        echo "$docker_token" | gh secret set DOCKER_TOKEN
        echo "✅ Docker Hub secrets configured"
        
        echo ""
        echo "📝 Note: You'll need to update .github/workflows/release.yml to use Docker Hub"
        echo "   Current setup publishes to: ghcr.io/grammatonic/pihole-network-analyzer"
    else
        echo "⏭️  Skipped Docker Hub setup"
    fi
fi

echo ""
echo "3. SLACK_WEBHOOK_URL (optional):"
echo "   - For release notifications to Slack"
echo ""

read -p "Do you want to set up Slack notifications? (y/N): " setup_slack

if [[ $setup_slack =~ ^[Yy]$ ]]; then
    echo ""
    echo "📋 To get a Slack webhook URL:"
    echo "1. Go to: https://api.slack.com/messaging/webhooks"
    echo "2. Create a new app or use existing"
    echo "3. Enable Incoming Webhooks"
    echo "4. Create webhook for your channel"
    echo "5. Copy the webhook URL"
    echo ""
    
    read -p "Enter Slack webhook URL: " -s slack_webhook
    echo ""
    
    if [ -n "$slack_webhook" ]; then
        echo "🔐 Setting SLACK_WEBHOOK_URL secret..."
        echo "$slack_webhook" | gh secret set SLACK_WEBHOOK_URL
        echo "✅ SLACK_WEBHOOK_URL secret configured"
    else
        echo "⏭️  Skipped Slack setup"
    fi
fi

echo ""
echo "📋 Final secrets configuration:"
gh secret list

echo ""
echo "🎉 Repository secrets configuration complete!"
echo ""
echo "📝 Summary:"
echo "- GITHUB_TOKEN: ✅ Automatically provided (sufficient for basic releases)"
echo "- Enhanced secrets: Configured as requested"
echo "- Ready for automated publishing: ✅"
echo ""
echo "🚀 Your release automation is now configured for:"
echo "   📦 GitHub Releases (automatic)"
echo "   🐳 GitHub Container Registry (ghcr.io)"
echo "   📊 Semantic versioning and changelog generation"
echo ""
echo "Next steps:"
echo "1. Test with: make release-dry-run"
echo "2. Create a release by merging to main or release/v* branch"
echo "3. Monitor the GitHub Actions workflow"
