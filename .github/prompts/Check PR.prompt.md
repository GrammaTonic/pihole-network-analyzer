---
mode: agent
---
Check the PR for this branch with gh api command and see what's going on and fix it, then git add commit push to trigger a new pipeline.
Monitor the new pipeline until it passes, if it errors analyse it and git add commit push to trigger a new one.
Continue like this until the pipeline is green.