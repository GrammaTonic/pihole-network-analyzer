#!/bin/bash
# Run these commands AFTER renaming the repository on GitHub

echo "🔄 Updating local git remote to new repository name..."
git remote set-url origin https://github.com/GrammaTonic/pihole-network-analyzer.git

echo "✅ Verifying new remote URL..."
git remote -v

echo "🚀 Testing connection to new repository..."
git push

echo "✅ Repository rename complete!"
echo "New repository URL: https://github.com/GrammaTonic/pihole-network-analyzer"
