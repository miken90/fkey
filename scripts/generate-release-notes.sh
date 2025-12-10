#!/bin/bash
# Generate release notes using AI (opencode CLI)
# Usage: ./generate-release-notes.sh [version] [from-ref]
# Examples:
#   ./generate-release-notes.sh                    # from last local tag to HEAD
#   ./generate-release-notes.sh v1.0.18            # from last local tag to HEAD
#   ./generate-release-notes.sh v1.0.18 v1.0.17   # from v1.0.17 to HEAD

VERSION="${1:-next}"
FROM_REF="$2"

# Determine starting point - prefer local tag (GitHub release may not exist yet)
if [ -z "$FROM_REF" ]; then
    # Get most recent tag from local git
    FROM_REF=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")
fi

# Fallback: get from GitHub release
if [ -z "$FROM_REF" ]; then
    FROM_REF=$(gh release view --json tagName -q .tagName 2>/dev/null || echo "")
fi

# Final fallback: last 20 commits
if [ -z "$FROM_REF" ]; then
    FROM_REF="HEAD~20"
fi

echo "📝 Generating release notes: $FROM_REF → HEAD" >&2

# Get commit list
COMMITS=$(git log "$FROM_REF"..HEAD --pretty=format:"- %s (%h)" 2>/dev/null)

# Get diff summary (files changed + stats)
DIFF_STAT=$(git diff "$FROM_REF"..HEAD --stat 2>/dev/null)

# Get detailed diff (limited to avoid being too long)
DIFF_CONTENT=$(git diff "$FROM_REF"..HEAD --no-color 2>/dev/null | head -500)

if [ -z "$COMMITS" ] && [ -z "$DIFF_STAT" ]; then
    echo "No changes found from $FROM_REF to HEAD" >&2
    exit 1
fi

echo "📊 Found $(echo "$COMMITS" | wc -l | tr -d ' ') commits" >&2

opencode run --format json "Tạo release notes cho version $VERSION của 'Gõ Nhanh' (Vietnamese IME for macOS).
Quy tắc:
- Phân tích code changes để hiểu thay đổi thực sự, không chỉ dựa vào commit message
- Nhóm theo: 🐛 Sửa lỗi, ⚡ Cải thiện, 🔧 Khác (bỏ section rỗng)
- Mỗi item: 1 dòng, súc tích, mô tả user-facing changes
- Viết tiếng Việt (có thể dùng keywords tiếng Anh như build, config, API...)
- Chỉ output markdown, không giải thích thêm" 2>/dev/null | jq -r 'select(.type == "text") | .part.text

## Commits:
$COMMITS

## Files changed:
$DIFF_STAT

## Code changes (snippet):
$DIFF_CONTENT
'
