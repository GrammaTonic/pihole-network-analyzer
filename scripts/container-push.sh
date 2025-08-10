#!/usr/bin/env bash
set -euo pipefail

# Pi-hole Network Analyzer - Container Publishing Script
# Comprehensive tool for building and pushing containers to GitHub Container Registry

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
REGISTRY="ghcr.io"
NAMESPACE="grammatonic"
IMAGE_NAME="pihole-analyzer"
FULL_IMAGE="${REGISTRY}/${NAMESPACE}/${IMAGE_NAME}"

# Get version information
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
COMMIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Platform support
PLATFORMS="linux/amd64,linux/arm64,linux/arm/v7"

# Function to print colored output
print_status() {
    echo -e "${BLUE}🚀 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${CYAN}💡 $1${NC}"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Docker is installed and running
    if ! docker --version >/dev/null 2>&1; then
        print_error "Docker is not installed or not running"
        exit 1
    fi
    
    # Check if buildx is available
    if ! docker buildx version >/dev/null 2>&1; then
        print_error "Docker buildx is not available"
        print_info "Please enable Docker buildx or update Docker to a newer version"
        exit 1
    fi
    
    # Check authentication
    if [ -z "${GITHUB_TOKEN:-}" ] && [ -z "${GITHUB_ACTOR:-}" ]; then
        print_warning "GITHUB_TOKEN and GITHUB_ACTOR not set"
        print_info "For GitHub Actions, these are set automatically"
        print_info "For local use, set them manually or use 'make docker-login-ghcr'"
    fi
    
    print_success "Prerequisites check passed"
}

# Function to setup buildx
setup_buildx() {
    print_status "Setting up Docker buildx for multi-architecture builds..."
    
    # Create builder instance if it doesn't exist
    if ! docker buildx ls | grep -q "multiarch"; then
        print_status "Creating multiarch builder instance..."
        docker buildx create --name multiarch --use --bootstrap
    else
        print_status "Using existing multiarch builder instance..."
        docker buildx use multiarch
    fi
    
    # Inspect the builder
    docker buildx inspect --bootstrap
    
    print_success "Buildx setup completed"
}

# Function to login to GHCR
login_ghcr() {
    print_status "Logging into GitHub Container Registry..."
    
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        echo "$GITHUB_TOKEN" | docker login ghcr.io -u "${GITHUB_ACTOR:-$(whoami)}" --password-stdin
        print_success "Successfully logged into GHCR"
    else
        print_warning "GITHUB_TOKEN not set, skipping login"
        print_info "Make sure you're already logged in with: docker login ghcr.io"
    fi
}

# Function to build and push production images
build_push_production() {
    print_status "Building and pushing production images to GHCR..."
    
    local tags=(
        "${FULL_IMAGE}:latest"
        "${FULL_IMAGE}:${VERSION}"
        "${FULL_IMAGE}:${COMMIT_SHA}"
    )
    
    # Add semantic version tags if VERSION is a proper semver
    if [[ $VERSION =~ ^v?[0-9]+\.[0-9]+\.[0-9]+.*$ ]]; then
        local clean_version="${VERSION#v}"
        local major=$(echo "$clean_version" | cut -d. -f1)
        local minor=$(echo "$clean_version" | cut -d. -f1-2)
        
        tags+=(
            "${FULL_IMAGE}:v${major}"
            "${FULL_IMAGE}:v${minor}"
        )
    fi
    
    # Build tag arguments
    local tag_args=""
    for tag in "${tags[@]}"; do
        tag_args="$tag_args -t $tag"
        print_info "Will tag as: $tag"
    done
    
    # Build and push
    docker buildx build \
        --platform "$PLATFORMS" \
        --target production \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        $tag_args \
        --push \
        --cache-from type=gha \
        --cache-to type=gha,mode=max \
        .
    
    print_success "Production images built and pushed successfully"
}

# Function to build and push development images
build_push_development() {
    print_status "Building and pushing development images to GHCR..."
    
    local dev_tags=(
        "${FULL_IMAGE}:dev"
        "${FULL_IMAGE}:dev-${VERSION}"
        "${FULL_IMAGE}:dev-${COMMIT_SHA}"
    )
    
    # Build tag arguments
    local tag_args=""
    for tag in "${dev_tags[@]}"; do
        tag_args="$tag_args -t $tag"
        print_info "Will tag as: $tag"
    done
    
    # Build and push development variant
    docker buildx build \
        --platform "$PLATFORMS" \
        --target development \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        $tag_args \
        --push \
        --cache-from type=gha \
        --cache-to type=gha,mode=max \
        .
    
    print_success "Development images built and pushed successfully"
}

# Function to list published images
list_images() {
    print_status "Listing published container images..."
    
    if command -v gh >/dev/null 2>&1; then
        print_info "Using GitHub CLI to list package versions..."
        gh api "/user/packages/container/${IMAGE_NAME}/versions" \
            --jq '.[] | select(.metadata.container.tags | length > 0) | {
                id: .id,
                tags: .metadata.container.tags,
                created: .created_at,
                updated: .updated_at,
                size: .package_size_bytes
            }' 2>/dev/null || print_warning "Failed to list images via GitHub API"
    else
        print_warning "GitHub CLI not found. Install with: brew install gh"
        print_info "You can view images at: https://github.com/GrammaTonic/pihole-network-analyzer/pkgs/container/pihole-analyzer"
    fi
}

# Function to show image information
show_image_info() {
    print_status "Image Information:"
    echo "Registry: $REGISTRY"
    echo "Namespace: $NAMESPACE"
    echo "Image Name: $IMAGE_NAME"
    echo "Full Image: $FULL_IMAGE"
    echo "Version: $VERSION"
    echo "Commit: $COMMIT_SHA"
    echo "Build Time: $BUILD_TIME"
    echo "Platforms: $PLATFORMS"
    echo ""
}

# Function to show usage
show_usage() {
    cat << EOF
Pi-hole Network Analyzer - Container Publishing Script

Usage: $0 [COMMAND]

Commands:
    production, prod    Build and push production images
    development, dev    Build and push development images
    all                 Build and push both production and development images
    list               List published container images
    info               Show image information
    login              Login to GitHub Container Registry
    help               Show this help message

Environment Variables:
    GITHUB_TOKEN       GitHub personal access token (required for pushing)
    GITHUB_ACTOR       GitHub username (defaults to current user)

Examples:
    $0 prod                    # Build and push production images
    $0 dev                     # Build and push development images
    $0 all                     # Build and push all images
    $0 list                    # List published images
    
    # With authentication
    export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
    export GITHUB_ACTOR=yourusername
    $0 prod

EOF
}

# Main function
main() {
    local command="${1:-help}"
    
    case "$command" in
        production|prod)
            show_image_info
            check_prerequisites
            setup_buildx
            login_ghcr
            build_push_production
            ;;
        development|dev)
            show_image_info
            check_prerequisites
            setup_buildx
            login_ghcr
            build_push_development
            ;;
        all)
            show_image_info
            check_prerequisites
            setup_buildx
            login_ghcr
            build_push_production
            build_push_development
            ;;
        list)
            list_images
            ;;
        info)
            show_image_info
            ;;
        login)
            check_prerequisites
            login_ghcr
            ;;
        help|--help|-h)
            show_usage
            ;;
        *)
            print_error "Unknown command: $command"
            echo ""
            show_usage
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
