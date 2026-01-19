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

# Force ASCII output to avoid encoding issues in WSL/terminal
$OutputEncoding = [System.Text.Encoding]::ASCII

# Auto-detect project root if not specified
if (-not $ProjectRoot) {
    # Script is at .claude/skills/release-github/scripts/github-release.ps1
    # Need to go up 4 levels to reach project root
    $ProjectRoot = (Get-Item $PSScriptRoot).Parent.Parent.Parent.Parent.FullName
}

$WailsDir = Join-Path $ProjectRoot "platforms\windows-wails"
$BuildScript = Join-Path $WailsDir "build.ps1"
$OutputDir = Join-Path $WailsDir "build\bin"
$TagName = "v$Version"

# Package name
$ZipName = "FKey-v$Version-portable.zip"
$ZipPath = Join-Path $OutputDir $ZipName

Write-Host "========================================"
Write-Host " FKey GitHub Release (Go/Wails)"
Write-Host "========================================"
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

# Step 2: Create ZIP package (single exe only)
Write-Host "[2/4] Creating ZIP package..." -ForegroundColor Yellow

$ExePath = Join-Path $OutputDir "FKey.exe"

if (-not (Test-Path $ExePath)) {
    throw "Executable not found: $ExePath"
}

# Create ZIP with single EXE (DLL is embedded)
if (Test-Path $ZipPath) {
    Remove-Item $ZipPath -Force
}

Compress-Archive -Path $ExePath -DestinationPath $ZipPath -CompressionLevel Optimal

$ZipSize = [math]::Round((Get-Item $ZipPath).Length / 1MB, 2)
Write-Host "[OK] Package ready: $ZipName ($ZipSize MB) - Single exe!" -ForegroundColor Green
Write-Host ""

# Step 3: Generate release notes
Write-Host "[3/4] Generating release notes..." -ForegroundColor Yellow

Push-Location $ProjectRoot
try {
    # Get previous tag (not the current one we're releasing)
    # First get all tags sorted by version, then find the one before current
    $AllTags = git tag --sort=-v:refname 2>$null
    $PrevTag = $null
    
    if ($AllTags) {
        $TagArray = $AllTags -split "`n" | Where-Object { $_ -and $_ -ne "v$Version" }
        if ($TagArray.Count -gt 0) {
            $PrevTag = $TagArray[0]
        }
    }

    # Get commits since previous tag (subject + body)
    # Use temp file because PowerShell doesn't preserve newlines in git output properly
    $tempFile = [System.IO.Path]::GetTempFileName()
    
    if ($PrevTag) {
        git log "$PrevTag..HEAD" --pretty=format:"---COMMIT---%n%s%n%b" --no-merges 2>$null | Out-File -FilePath $tempFile -Encoding UTF8
        $CompareLink = "**Full Changelog**: https://github.com/$Repo/compare/$PrevTag...v$Version"
    } else {
        git log -20 --pretty=format:"---COMMIT---%n%s%n%b" --no-merges 2>$null | Out-File -FilePath $tempFile -Encoding UTF8
        $CompareLink = ""
    }
    
    $CommitData = Get-Content $tempFile -Raw -ErrorAction SilentlyContinue
    Remove-Item $tempFile -ErrorAction SilentlyContinue

    # Categorize commits
    $Features = @()
    $Fixes = @()
    $Improvements = @()

    if ($CommitData) {
        # Normalize line endings
        $CommitData = $CommitData -replace "`r`n", "`n"
        
        # Split by commit delimiter
        $Commits = $CommitData -split "---COMMIT---" | Where-Object { $_.Trim() }
        
        foreach ($commitBlock in $Commits) {
            $lines = $commitBlock.Trim() -split "`n"
            if ($lines.Count -eq 0) { continue }
            
            $subject = $lines[0].Trim()
            $bodyLines = @()
            if ($lines.Count -gt 1) {
                $bodyLines = $lines[1..($lines.Count-1)] | ForEach-Object { $_.Trim() } | Where-Object { $_ -match "^-\s" }
            }
            
            if (-not $subject) { continue }
            
            # Skip merge commits and version-only commits
            if ($subject -match "^Merge" -or $subject -match "^v\d+\.\d+\.\d+$") { continue }
            
            # Filter by platform prefix: only include [win], [core], [all] for Windows releases
            # Skip [linux] commits
            if ($subject -match "^\[linux\]") { continue }
            
            # Remove platform prefix for display
            $subject = $subject -replace "^\[(win|core|all)\]\s*", ""
            
            # Determine default category from subject
            $defaultCategory = "improvements"
            if ($subject -match "^feat(\(.+\))?:|^feature:|^add:|^new:") {
                $defaultCategory = "features"
            }
            elseif ($subject -match "^fix(\(.+\))?:|^bug:|^hotfix:") {
                $defaultCategory = "fixes"
            }
            
            # If body has bullet points, parse each line for its own category
            if ($bodyLines.Count -gt 0) {
                foreach ($line in $bodyLines) {
                    # Remove leading dash and spaces
                    $item = $line -replace "^-\s*", ""
                    $item = $item.Trim()
                    if (-not $item) { continue }
                    
                    # Check if line itself has a category prefix
                    $lineCategory = $defaultCategory
                    if ($item -match "^feat(\(.+\))?:|^feature:") {
                        $lineCategory = "features"
                        $item = $item -replace "^feat(\(.+\))?:\s*", ""
                        $item = $item -replace "^feature:\s*", ""
                    }
                    elseif ($item -match "^fix(\(.+\))?:|^bug:") {
                        $lineCategory = "fixes"
                        $item = $item -replace "^fix(\(.+\))?:\s*", ""
                        $item = $item -replace "^bug:\s*", ""
                    }
                    
                    $item = $item.Trim()
                    if ($item) {
                        switch ($lineCategory) {
                            "features" { $Features += "- $item" }
                            "fixes" { $Fixes += "- $item" }
                            default { $Improvements += "- $item" }
                        }
                    }
                }
            }
            else {
                # Clean up prefix for display
                $msg = $subject -replace "^(feat|feature|fix|bug|hotfix|chore|refactor|perf|docs|test|ci|build|add|new)(\(.+\))?:\s*", ""
                if ($msg) {
                    switch ($defaultCategory) {
                        "features" { $Features += "- $msg" }
                        "fixes" { $Fixes += "- $msg" }
                        default { $Improvements += "- $msg" }
                    }
                }
            }
        }
    }

    # Build release notes sections (ASCII only - no emoji to avoid encoding issues)
    $Sections = @()
    
    if ($Features.Count -gt 0) {
        $Sections += "### New Features`n`n" + ($Features -join "`n")
    }
    
    if ($Fixes.Count -gt 0) {
        $Sections += "### Bug Fixes`n`n" + ($Fixes -join "`n")
    }
    
    if ($Improvements.Count -gt 0) {
        $Sections += "### Improvements`n`n" + ($Improvements -join "`n")
    }

    # Fallback if no categorized commits
    if ($Sections.Count -eq 0) {
        $Sections += "- Initial release"
    }

    $ChangesSection = $Sections -join "`n`n"

    # Write release notes to temp file
    $NotesFile = Join-Path $env:TEMP "release-notes-$Version.md"
    
    $ReleaseNotes = @"
## What's Changed

$ChangesSection

---

## Download

| Platform | File | Size |
|----------|------|------|
| Windows (Portable) | [$ZipName](https://github.com/$Repo/releases/download/v$Version/$ZipName) | ~$ZipSize MB |

### Installation

1. Download and extract ``$ZipName``
2. Run ``FKey.exe`` (single file, no DLL needed)
3. App runs in system tray

$CompareLink
"@

    # Write with UTF-8 no BOM
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
Write-Host "========================================"
Write-Host " Release Complete"
Write-Host "========================================"
Write-Host "Version:  $Version" -ForegroundColor White
Write-Host "Tag:      $TagName" -ForegroundColor White
Write-Host "Package:  $ZipSize MB" -ForegroundColor White

$ReleaseUrl = "https://github.com/$Repo/releases/tag/$TagName"
Write-Host "GitHub:   $ReleaseUrl" -ForegroundColor White

Write-Host ""
Write-Host "[SUCCESS] Release published!" -ForegroundColor Green
Write-Host ""
