#!/bin/bash
# Generate release notes using AI (opencode CLI)

LAST_RELEASE=$(gh release view --json tagName -q .tagName 2>/dev/null || echo "")
VERSION="$1"

if [ -n "$LAST_RELEASE" ]; then
    COMMITS=$(git log "$LAST_RELEASE"..HEAD --pretty=format:"%s" 2>/dev/null)
else
    COMMITS=$(git log --pretty=format:"%s" -20 2>/dev/null)
fi

if [ -z "$COMMITS" ]; then
    echo "No commits found"
    exit 1
fi

opencode run --format json "Generate concise release notes for version $VERSION of 'Gõ Nhanh' (Vietnamese IME for macOS).

Commits:
$COMMITS

Rules:
- Group by: Features, Fixes, Improvements, Other
- Skip empty sections
- Each item: one line, start with emoji (✨ feat, 🐛 fix, ⚡ perf, 📝 docs, 🔧 chore)
- Be concise, no fluff
- Output markdown only, no explanation" 2>/dev/null | jq -r 'select(.type == "text") | .content' | tr -d '\n' | sed 's/$/\n/'
