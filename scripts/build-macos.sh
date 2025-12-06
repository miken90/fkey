#!/bin/bash
set -e

echo "🍎 Building macOS app..."

# Build core first
./scripts/build-core.sh

# Build macOS app with xcodebuild
cd "$(dirname "$0")/../platforms/macos"

if [ -d "GoNhanh.xcodeproj" ]; then
    echo "Building with Xcode..."
    xcodebuild -scheme GoNhanh -configuration Release
    echo "✅ macOS app built successfully!"
    echo "📦 App: platforms/macos/build/Release/GoNhanh.app"
else
    echo "⚠️  Xcode project not found!"
    echo "Please create Xcode project in platforms/macos/"
    echo "Steps:"
    echo "  1. Open Xcode"
    echo "  2. Create new macOS App project"
    echo "  3. Name: GoNhanh"
    echo "  4. Location: platforms/macos/"
    echo "  5. Add Swift files from GoNhanh/ folder"
    echo "  6. Link libgonhanh_core.a in Build Phases"
fi
