# FKey - GitHub Release Script
# Builds and uploads release to GitHub Releases using Go/Wails build
# Usage: .\github-release.ps1 -Version "2.0.0"
# Note: Only pushes to origin (miken90), never to upstream

param(
    [Parameter(Mandatory=$true)]
    [string]$Version,

    [string]$ProjectRoot = "",
    [string]$Repo = "miken90/fkey",
    [switch]$SkipBuild,
    [switch]$Draft,
    [switch]$Sign
)

$ErrorActionPreference = "Stop"
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$PSDefaultParameterValues['Out-File:Encoding'] = 'utf8'

# Auto-detect project root if not specified
if (-not $ProjectRoot) {
    $ProjectRoot = (Get-Item $PSScriptRoot).Parent.Parent.Parent.FullName
    # Fallback: look for gonhanh.org in common locations
    if (-not (Test-Path "$ProjectRoot\platforms\windows-wails\build.ps1")) {
        $ProjectRoot = "C:\WORKSPACES\2026\gonhanh.org"
    }
}

$WailsDir = Join-Path $ProjectRoot "platforms\windows-wails"
$BuildScript = Join-Path $WailsDir "build.ps1"
$OutputDir = Join-Path $WailsDir "build\bin"
$TagName = "v$Version"

# Package name
$ZipName = "FKey-v$Version-portable.zip"
$ZipPath = Join-Path $OutputDir $ZipName

Write-Host "════════════════════════════════════════" -ForegroundColor Cyan
Write-Host " FKey GitHub Release (Go/Wails)" -ForegroundColor Cyan
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
    Write-Host "[1/4] Building portable package..." -ForegroundColor Yellow
    Write-Host ""

    Push-Location $WailsDir
    try {
        $buildArgs = @("-Release", "-Version", $Version)
        if ($Sign) {
            $buildArgs += "-Sign"
        }
        & $BuildScript @buildArgs
        if ($LASTEXITCODE -ne 0) { throw "Build failed" }
    }
    finally {
        Pop-Location
    }

    Write-Host ""
}
else {
    Write-Host "[1/4] Skipping build (--SkipBuild)" -ForegroundColor Gray
    Write-Host ""
}

# Step 2: Create ZIP package
Write-Host "[2/4] Creating ZIP package..." -ForegroundColor Yellow

$ExePath = Join-Path $OutputDir "FKey.exe"
$DllPath = Join-Path $OutputDir "gonhanh_core.dll"

if (-not (Test-Path $ExePath)) {
    throw "Executable not found: $ExePath"
}

# Create ZIP with both EXE and DLL
if (Test-Path $ZipPath) {
    Remove-Item $ZipPath -Force
}

Compress-Archive -Path $ExePath, $DllPath -DestinationPath $ZipPath -CompressionLevel Optimal

$ZipSize = [math]::Round((Get-Item $ZipPath).Length / 1MB, 2)
Write-Host "[OK] Package ready: $ZipName ($ZipSize MB)" -ForegroundColor Green
Write-Host ""

# Step 3: Generate release notes
Write-Host "[3/4] Generating release notes..." -ForegroundColor Yellow

Push-Location $ProjectRoot
try {
    # Get last tag
    $LastTag = git describe --tags --abbrev=0 2>$null

    # Get commits since last tag
    if ($LastTag) {
        $CommitLines = git log "$LastTag..HEAD" --pretty=format:"%s" --no-merges 2>$null
        $CompareLink = "**Full Changelog**: https://github.com/$Repo/compare/$LastTag...v$Version"
    } else {
        $CommitLines = git log -20 --pretty=format:"%s" --no-merges 2>$null
        $CompareLink = ""
    }

    # Categorize commits
    $Features = @()
    $Fixes = @()
    $Improvements = @()

    if ($CommitLines) {
        $CommitArray = $CommitLines -split "`n"
        foreach ($commit in $CommitArray) {
            $commit = $commit.Trim()
            if (-not $commit) { continue }
            
            # Skip merge commits and version commits
            if ($commit -match "^Merge" -or $commit -match "^v\d+\.\d+") { continue }
            
            # Categorize by conventional commit prefix
            if ($commit -match "^feat(\(.+\))?:|^feature:|^add:|^new:") {
                # Clean up prefix for display
                $msg = $commit -replace "^feat(\(.+\))?:\s*", ""
                $msg = $msg -replace "^feature:\s*", ""
                $msg = $msg -replace "^add:\s*", ""
                $msg = $msg -replace "^new:\s*", ""
                if ($msg) { $Features += "- $msg" }
            }
            elseif ($commit -match "^fix(\(.+\))?:|^bug:|^hotfix:") {
                $msg = $commit -replace "^fix(\(.+\))?:\s*", ""
                $msg = $msg -replace "^bug:\s*", ""
                $msg = $msg -replace "^hotfix:\s*", ""
                if ($msg) { $Fixes += "- $msg" }
            }
            elseif ($commit -match "^refactor:|^perf:|^improve:|^chore:|^style:|^docs:|^test:|^ci:|^build:") {
                $msg = $commit -replace "^(refactor|perf|improve|chore|style|docs|test|ci|build)(\(.+\))?:\s*", ""
                if ($msg) { $Improvements += "- $msg" }
            }
            else {
                # Uncategorized commits go to improvements
                $Improvements += "- $commit"
            }
        }
    }

    # Build release notes sections
    $Sections = @()
    
    if ($Features.Count -gt 0) {
        $Sections += "### ✨ New Features`n`n" + ($Features -join "`n")
    }
    
    if ($Fixes.Count -gt 0) {
        $Sections += "### 🐛 Bug Fixes`n`n" + ($Fixes -join "`n")
    }
    
    if ($Improvements.Count -gt 0) {
        $Sections += "### ⚡ Improvements`n`n" + ($Improvements -join "`n")
    }

    # Fallback if no categorized commits
    if ($Sections.Count -eq 0) {
        $Sections += "- Initial release"
    }

    $ChangesSection = $Sections -join "`n`n"

    # Write release notes to temp file (fixes UTF-8 encoding issues)
    $NotesFile = Join-Path $env:TEMP "release-notes-$Version.md"
    
    $ReleaseNotes = @"
## What's Changed

$ChangesSection

---

## 📦 Download

| Platform | File | Size |
|----------|------|------|
| Windows (Portable) | [$ZipName](https://github.com/$Repo/releases/download/v$Version/$ZipName) | ~$ZipSize MB |

### Installation

1. Download và giải nén ``$ZipName``
2. Chạy ``FKey.exe``
3. Ứng dụng chạy ở khay hệ thống (system tray)

$CompareLink
"@

    # Write with UTF-8 BOM for proper encoding
    [System.IO.File]::WriteAllText($NotesFile, $ReleaseNotes, [System.Text.UTF8Encoding]::new($false))

    Write-Host "[OK] Release notes generated" -ForegroundColor Green
    Write-Host ""
}
finally {
    Pop-Location
}

# Step 4: Create GitHub release
Write-Host "[4/4] Creating GitHub release..." -ForegroundColor Yellow

Push-Location $ProjectRoot
try {
    Write-Host "  Repo: $Repo" -ForegroundColor Gray
    Write-Host "  Tag: $TagName" -ForegroundColor Gray
    Write-Host "  Asset: $ZipName" -ForegroundColor Gray
    Write-Host ""

    # Push code and tag to origin only (never upstream)
    $prevErrorAction = $ErrorActionPreference
    $ErrorActionPreference = "Continue"

    Write-Host "  Pushing code to origin..." -ForegroundColor Gray
    $null = git push origin HEAD 2>&1
    
    Write-Host "  Creating and pushing tag..." -ForegroundColor Gray
    $null = git tag $TagName 2>&1
    $null = git push origin $TagName 2>&1

    $ErrorActionPreference = $prevErrorAction

    # Create release with gh CLI using notes file
    $ReleaseArgs = @(
        "release", "create", $TagName, $ZipPath,
        "--repo", $Repo,
        "--title", "FKey v$Version",
        "--notes-file", $NotesFile
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

    # Clean up temp file
    Remove-Item $NotesFile -ErrorAction SilentlyContinue

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

$ReleaseUrl = "https://github.com/$Repo/releases/tag/$TagName"
Write-Host "GitHub:   $ReleaseUrl" -ForegroundColor White

Write-Host ""
Write-Host "[SUCCESS] Release published!" -ForegroundColor Green
Write-Host ""
