#!/bin/bash
# Generate release notes using AI (opencode CLI)
# Usage: ./generate-release-notes.sh [version] [from-commit]
# Examples:
#   ./generate-release-notes.sh                    # từ last release đến HEAD
#   ./generate-release-notes.sh v1.0.18            # từ last release đến HEAD, version v1.0.18
#   ./generate-release-notes.sh v1.0.18 abc123     # từ commit abc123 đến HEAD

VERSION="${1:-next}"
FROM_COMMIT="$2"

if [ -n "$FROM_COMMIT" ]; then
    # Từ commit được chỉ định
    COMMITS=$(git log "$FROM_COMMIT"..HEAD --pretty=format:"%s" 2>/dev/null)
else
    # Từ last GitHub release
    LAST_RELEASE=$(gh release view --json tagName -q .tagName 2>/dev/null || echo "")
    if [ -n "$LAST_RELEASE" ]; then
        COMMITS=$(git log "$LAST_RELEASE"..HEAD --pretty=format:"%s" 2>/dev/null)
    fi
fi

# Fallback: 20 commits gần nhất
if [ -z "$COMMITS" ]; then
    COMMITS=$(git log --pretty=format:"%s" -20 2>/dev/null)
fi

if [ -z "$COMMITS" ]; then
    echo "Không tìm thấy commits"
    exit 1
fi

opencode run --format json "Tạo release notes cho version $VERSION của 'Gõ Nhanh' (Vietnamese IME for macOS).

Commits:
$COMMITS

Quy tắc:
- Nhóm theo: ✨ Tính năng mới, 🐛 Sửa lỗi, ⚡ Cải thiện, 🔧 Khác
- Bỏ qua section rỗng
- Mỗi item: 1 dòng, súc tích, viết tiếng Việt (có thể dùng keywords tiếng Anh như build, config, API...)
- Chỉ output markdown, không giải thích" 2>/dev/null | jq -r 'select(.type == "text") | .part.text'
