#!/bin/bash

# Cache warming script for faster subsequent builds
# This script pre-downloads and caches dependencies

set -euo pipefail

echo "🔥 Warming up build caches..."

# Start timing
start_time=$(date +%s)

# Warm up Go module cache
echo "📦 Downloading Go modules..."
go mod download

# Warm up build cache by building core packages
echo "🏗️  Pre-building core packages..."
WARM_DIRS=(
    "./cmd"
    "./internal/analyzer"
    "./internal/cli"
    "./internal/colors"
    "./internal/config"
    "./internal/interfaces"
    "./internal/logger"
    "./internal/network"
    "./internal/pihole"
    "./internal/reporting"
    "./internal/types"
)for pkg in "${packages[@]}"; do
    echo "  Building $pkg..."
    go build -i "$pkg" >/dev/null 2>&1 || true
done

# Pre-compile test packages
echo "🧪 Pre-building test packages..."
go test -i ./... >/dev/null 2>&1 || true

# Calculate timing
end_time=$(date +%s)
duration=$((end_time - start_time))

echo "✅ Cache warming completed in ${duration}s"
echo "🚀 Subsequent builds will be faster!"
