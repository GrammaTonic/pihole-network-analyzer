#!/bin/bash

# Pre-push Test Script
# Run this before pushing to ensure CI will pass

set -e

echo "🧪 Pre-push Validation Script"
echo "=============================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
        exit 1
    fi
}

print_info() {
    echo -e "${YELLOW}🔍 $1${NC}"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}❌ Error: go.mod not found. Run this script from the project root.${NC}"
    exit 1
fi

print_info "Step 1: Checking Go modules..."
go mod tidy
go mod verify
print_status $? "Go modules verified"

print_info "Step 2: Downloading dependencies..."
go mod download
print_status $? "Dependencies downloaded"

print_info "Step 3: Checking code formatting..."
UNFORMATTED=$(gofmt -s -l . | wc -l)
if [ "$UNFORMATTED" -gt 0 ]; then
    echo -e "${RED}❌ Code formatting issues found:${NC}"
    gofmt -s -l .
    echo ""
    echo -e "${YELLOW}💡 Fix with: gofmt -s -w .${NC}"
    exit 1
fi
print_status 0 "Code formatting is correct"

print_info "Step 4: Running go vet..."
go vet ./...
print_status $? "Static analysis passed"

print_info "Step 5: Building application..."
go build -o pihole-analyzer ./cmd/pihole-analyzer
print_status $? "Application built successfully"

print_info "Step 6: Running test suite..."
./pihole-analyzer --test
print_status $? "All tests passed"

print_info "Step 7: Testing multi-platform builds..."
echo "  - Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -o /tmp/test-linux-amd64 ./cmd/pihole-analyzer
print_status $? "Linux AMD64 build"

echo "  - Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -o /tmp/test-windows-amd64.exe ./cmd/pihole-analyzer
print_status $? "Windows AMD64 build"

echo "  - macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -o /tmp/test-darwin-arm64 ./cmd/pihole-analyzer
print_status $? "macOS ARM64 build"

# Cleanup test builds
rm -f /tmp/test-linux-amd64 /tmp/test-windows-amd64.exe /tmp/test-darwin-arm64

print_info "Step 8: Checking for security vulnerabilities..."
if command -v govulncheck &> /dev/null; then
    govulncheck ./... || echo -e "${YELLOW}⚠️ Vulnerabilities found (CI will continue)${NC}"
    print_status 0 "Security scan completed"
else
    echo -e "${YELLOW}⚠️ govulncheck not installed, skipping security scan${NC}"
    echo -e "${YELLOW}💡 Install with: go install golang.org/x/vuln/cmd/govulncheck@latest${NC}"
fi

# Show current branch info
CURRENT_BRANCH=$(git branch --show-current)
echo ""
echo "📋 Current Status:"
echo "   Branch: $CURRENT_BRANCH"

if [ "$CURRENT_BRANCH" = "main" ] || [ "$CURRENT_BRANCH" = "master" ]; then
    echo -e "   ${YELLOW}⚠️  You're on the main branch - builds will be created!${NC}"
else
    echo -e "   ${GREEN}✅ Feature branch - only build verification will run${NC}"
fi

echo ""
echo -e "${GREEN}🎉 All checks passed! Your code is ready to push.${NC}"
echo ""
echo "Next steps:"
echo "  git add ."
echo "  git commit -m 'Your commit message'"
echo "  git push origin $CURRENT_BRANCH"

# Cleanup
rm -f pihole-analyzer
