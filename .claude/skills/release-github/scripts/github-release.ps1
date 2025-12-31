# GoNhanh - GitHub Release Script
# Builds and uploads release to GitHub Releases using standardized build script
# Usage: .\github-release.ps1 -Version "1.5.9"
# Note: Only pushes to origin (miken90), never to upstream

param(
    [Parameter(Mandatory=$true)]
    [string]$Version,

    [string]$ProjectRoot = "",
    [string]$Repo = "miken90/gonhanh.org",  # Target repo for release (never upstream)
    [switch]$SkipBuild,
    [switch]$Draft
)

$ErrorActionPreference = "Stop"

# Auto-detect project root if not specified
if (-not $ProjectRoot) {
    $ProjectRoot = (Get-Item $PSScriptRoot).Parent.Parent.Parent.FullName
    # Fallback: look for gonhanh.org in common locations
    if (-not (Test-Path "$ProjectRoot\platforms\windows\build-release.ps1")) {
        $ProjectRoot = "C:\WORKSPACES\PERSONAL\gonhanh.org"
    }
}

$WindowsDir = Join-Path $ProjectRoot "platforms\windows"
$BuildScript = Join-Path $WindowsDir "build-release.ps1"
$PublishDir = Join-Path $WindowsDir "GoNhanh\bin\Release\net8.0-windows\win-x64\publish"
$ZipName = "FKey-v$Version-portable.zip"
$ZipPath = Join-Path $PublishDir $ZipName
$TagName = "v$Version"

Write-Host "════════════════════════════════════════" -ForegroundColor Cyan
Write-Host " FKey GitHub Release" -ForegroundColor Cyan
Write-Host "════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "Version:  $Version" -ForegroundColor White
Write-Host "Tag:      $TagName" -ForegroundColor White
Write-Host "Project:  $ProjectRoot" -ForegroundColor White
Write-Host ""

# Verify gh CLI is available
if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    throw "GitHub CLI (gh) not found. Install from https://cli.github.com/"
}

# Verify gh is authenticated
$ghAuth = gh auth status 2>&1
if ($LASTEXITCODE -ne 0) {
    throw "GitHub CLI not authenticated. Run: gh auth login"
}

# Verify build script exists
if (-not (Test-Path $BuildScript)) {
    throw "Build script not found: $BuildScript"
}

# Step 1: Build portable package
if (-not $SkipBuild) {
    Write-Host "[1/3] Building portable package..." -ForegroundColor Yellow
    Write-Host ""

    Push-Location $WindowsDir
    try {
        & $BuildScript -Version $Version
        if ($LASTEXITCODE -ne 0) { throw "Build failed" }
    }
    finally {
        Pop-Location
    }

    Write-Host ""
}
else {
    Write-Host "[1/3] Skipping build (--SkipBuild)" -ForegroundColor Gray
    Write-Host ""
}

# Verify ZIP exists
if (-not (Test-Path $ZipPath)) {
    throw "Build artifact not found: $ZipPath"
}

$ZipSize = [math]::Round((Get-Item $ZipPath).Length / 1MB, 2)
Write-Host "Package ready: $ZipName ($ZipSize MB)" -ForegroundColor Green
Write-Host ""

# Step 2: Generate release notes
Write-Host "[2/3] Generating release notes..." -ForegroundColor Yellow

Push-Location $ProjectRoot
try {
    # Get last tag
    $LastTag = git describe --tags --abbrev=0 2>$null

    # Generate notes from commits
    if ($LastTag) {
        $Commits = git log "$LastTag..HEAD" --pretty=format:"- %s" --no-merges 2>$null
        $CompareLink = "**Full Changelog**: https://github.com/$Repo/compare/$LastTag...v$Version"
    } else {
        # No previous tags, get last 10 commits
        $Commits = git log -10 --pretty=format:"- %s" --no-merges 2>$null
        $CompareLink = ""
    }

    if (-not $Commits) {
        $Commits = "- Initial release"
    }

    $ReleaseNotes = @"
## What's New in v$Version

$Commits

## Download

- **Windows Portable**: [FKey-v$Version-portable.zip](https://github.com/$Repo/releases/download/v$Version/$ZipName) (~$ZipSize MB)

## Installation

1. Download ``FKey-v$Version-portable.zip``
2. Extract and run ``FKey.exe``
3. App runs in system tray

$CompareLink
"@

    Write-Host "[OK] Release notes generated" -ForegroundColor Green
    Write-Host ""
}
finally {
    Pop-Location
}

# Step 3: Create GitHub release
Write-Host "[3/3] Creating GitHub release..." -ForegroundColor Yellow

Push-Location $ProjectRoot
try {
    Write-Host "  Repo: $Repo" -ForegroundColor Gray
    Write-Host "  Tag: $TagName" -ForegroundColor Gray
    Write-Host "  Asset: $ZipName" -ForegroundColor Gray
    Write-Host ""

    # Create tag locally and push to origin only (never upstream)
    # Temporarily allow errors for git commands (stderr output is not actual errors)
    $prevErrorAction = $ErrorActionPreference
    $ErrorActionPreference = "Continue"

    $null = git tag $TagName 2>&1
    $null = git push origin $TagName 2>&1

    $ErrorActionPreference = $prevErrorAction

    # Create release with gh CLI - explicitly specify repo to avoid pushing to wrong remote
    $ReleaseArgs = @(
        "release", "create", $TagName, $ZipPath,
        "--repo", $Repo,
        "--title", "FKey v$Version",
        "--notes", $ReleaseNotes
    )
    if ($Draft) {
        $ReleaseArgs += "--draft"
    }

    # Clear GITHUB_TOKEN to use gh's stored credentials
    $env:GH_TOKEN = ""
    & gh @ReleaseArgs

    if ($LASTEXITCODE -ne 0) {
        throw "Failed to create GitHub release"
    }

    Write-Host "[OK] Release created" -ForegroundColor Green
}
finally {
    Pop-Location
}

# Summary
Write-Host ""
Write-Host "════════════════════════════════════════" -ForegroundColor Cyan
Write-Host " Release Complete" -ForegroundColor Cyan
Write-Host "════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "Version:  $Version" -ForegroundColor White
Write-Host "Tag:      $TagName" -ForegroundColor White
Write-Host "Package:  $ZipSize MB" -ForegroundColor White

# Show release URL (use $Repo directly, not gh repo view which may pick wrong repo)
$ReleaseUrl = "https://github.com/$Repo/releases/tag/$TagName"
Write-Host "GitHub:   $ReleaseUrl" -ForegroundColor White

Write-Host ""
Write-Host "[SUCCESS] Release published!" -ForegroundColor Green
Write-Host ""
