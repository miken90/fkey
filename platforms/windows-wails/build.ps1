# FKey Wails v3 Build Script
# Run from Windows PowerShell

param(
    [switch]$Release,
    [switch]$Dev,
    [switch]$Clean,
    [switch]$Sign,
    [switch]$CreateCert,
    [string]$CertPath = "",
    [string]$CertPassword = "",
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"
$ProjectDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$CoreDir = Join-Path (Split-Path -Parent (Split-Path -Parent $ProjectDir)) "core"

Write-Host "FKey Build Script" -ForegroundColor Cyan
Write-Host "=================" -ForegroundColor Cyan

# Create self-signed certificate for testing
if ($CreateCert) {
    Write-Host "`nCreating self-signed code signing certificate..." -ForegroundColor Yellow
    $cert = New-SelfSignedCertificate -Type CodeSigningCert `
        -Subject "CN=FKey Vietnamese IME, O=GoNhanh" `
        -KeyAlgorithm RSA -KeyLength 2048 `
        -CertStoreLocation Cert:\CurrentUser\My `
        -NotAfter (Get-Date).AddYears(3)
    
    Write-Host "Certificate created:" -ForegroundColor Green
    Write-Host "  Subject: $($cert.Subject)" -ForegroundColor Cyan
    Write-Host "  Thumbprint: $($cert.Thumbprint)" -ForegroundColor Cyan
    Write-Host "  Expires: $($cert.NotAfter)" -ForegroundColor Cyan
    Write-Host "`nNote: Self-signed certs still trigger SmartScreen warning." -ForegroundColor Yellow
    Write-Host "For trusted signing, purchase a certificate from DigiCert/Sectigo." -ForegroundColor DarkGray
    exit 0
}

# Get version from git tag if not specified
function Get-GitVersion {
    try {
        $tag = git describe --tags --abbrev=0 2>$null
        if ($tag) {
            # Remove 'v' prefix if present
            return $tag -replace '^v', ''
        }
    } catch {}
    
    # Fallback: try git describe with commit hash
    try {
        $desc = git describe --tags --always 2>$null
        if ($desc) {
            return $desc -replace '^v', ''
        }
    } catch {}
    
    return "dev"
}

# Code signing function
function Sign-Binary {
    param([string]$FilePath)
    
    if (-not $Sign) { return }
    
    # Find signtool.exe
    $SignTool = $null
    $SDKPaths = @(
        "${env:ProgramFiles(x86)}\Windows Kits\10\bin\*\x64\signtool.exe",
        "${env:ProgramFiles}\Windows Kits\10\bin\*\x64\signtool.exe"
    )
    foreach ($pattern in $SDKPaths) {
        $found = Get-Item $pattern -ErrorAction SilentlyContinue | Sort-Object -Descending | Select-Object -First 1
        if ($found) { $SignTool = $found.FullName; break }
    }
    
    if (-not $SignTool) {
        Write-Host "WARNING: signtool.exe not found. Install Windows SDK." -ForegroundColor Yellow
        Write-Host "  winget install Microsoft.WindowsSDK" -ForegroundColor DarkGray
        return
    }
    
    # Determine certificate source
    if ($CertPath -and (Test-Path $CertPath)) {
        # Use provided PFX certificate
        Write-Host "Signing with certificate: $CertPath" -ForegroundColor Yellow
        $signArgs = @("sign", "/f", $CertPath, "/fd", "SHA256", "/tr", "http://timestamp.digicert.com", "/td", "SHA256")
        if ($CertPassword) {
            $signArgs += @("/p", $CertPassword)
        }
        $signArgs += $FilePath
        & $SignTool @signArgs
    }
    else {
        # Try to use certificate from Windows Certificate Store
        $cert = Get-ChildItem -Path Cert:\CurrentUser\My -CodeSigningCert | Where-Object { $_.Subject -like "*FKey*" -or $_.Subject -like "*GoNhanh*" } | Select-Object -First 1
        
        if ($cert) {
            Write-Host "Signing with certificate: $($cert.Subject)" -ForegroundColor Yellow
            & $SignTool sign /sha1 $cert.Thumbprint /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 $FilePath
        }
        else {
            Write-Host "No signing certificate found." -ForegroundColor Yellow
            Write-Host "Options:" -ForegroundColor DarkGray
            Write-Host "  1. Create self-signed: .\build.ps1 -Release -Sign -CreateCert" -ForegroundColor DarkGray
            Write-Host "  2. Use PFX file: .\build.ps1 -Release -Sign -CertPath 'cert.pfx' -CertPassword 'xxx'" -ForegroundColor DarkGray
            Write-Host "  3. Purchase from DigiCert, Sectigo, Comodo (~\$200-500/year)" -ForegroundColor DarkGray
            return
        }
    }
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Signed: $FilePath" -ForegroundColor Green
    } else {
        Write-Host "WARNING: Signing failed for $FilePath" -ForegroundColor Red
    }
}

# Change to project directory
Push-Location $ProjectDir

try {
    if ($Clean) {
        Write-Host "`nCleaning build artifacts..." -ForegroundColor Yellow
        Remove-Item -Path "build\bin\*" -Recurse -Force -ErrorAction SilentlyContinue
        Remove-Item -Path "FKey.exe" -Force -ErrorAction SilentlyContinue
        Write-Host "Done." -ForegroundColor Green
        exit 0
    }

    # Determine version
    if (-not $Version) {
        $Version = Get-GitVersion
    }
    Write-Host "`nVersion: $Version" -ForegroundColor Magenta

    # Check if Rust DLL exists - build/update if needed
    $DllSource = Join-Path $CoreDir "target\release\gonhanh_core.dll"
    $DllDest = Join-Path $ProjectDir "gonhanh_core.dll"
    
    # Always rebuild DLL for release builds, or if missing
    $NeedBuildDll = $Release -or !(Test-Path $DllDest)
    if ($NeedBuildDll) {
        Write-Host "`nBuilding Rust DLL..." -ForegroundColor Yellow
        Push-Location $CoreDir
        cargo build --release
        Pop-Location
        
        if (Test-Path $DllSource) {
            Copy-Item $DllSource $DllDest -Force
            Write-Host "DLL copied to project directory (for embedding)." -ForegroundColor Green
        } else {
            Write-Host "ERROR: Failed to build Rust DLL" -ForegroundColor Red
            exit 1
        }
    }

    # Get dependencies
    Write-Host "`nGetting Go dependencies..." -ForegroundColor Yellow
    go mod tidy

    # Generate Windows resources (.syso) with icon
    Write-Host "`nGenerating Windows resources..." -ForegroundColor Yellow
    $GoWinRes = Get-Command go-winres -ErrorAction SilentlyContinue
    if ($GoWinRes) {
        # Update version in winres.json if specified
        if ($Version -and $Version -ne "dev") {
            $winresFile = Join-Path $ProjectDir "winres.json"
            if (Test-Path $winresFile) {
                $winresContent = Get-Content $winresFile -Raw | ConvertFrom-Json
                $versionParts = $Version -split '\.'
                $versionFull = "$($versionParts[0]).$($versionParts[1]).$($versionParts[2]).0"
                $winresContent.RT_VERSION.'#1'.'0409'.fixed.file_version = $versionFull
                $winresContent.RT_VERSION.'#1'.'0409'.fixed.product_version = $versionFull
                $winresContent.RT_VERSION.'#1'.'0409'.info.'0409'.FileVersion = $Version
                $winresContent.RT_VERSION.'#1'.'0409'.info.'0409'.ProductVersion = $Version
                $winresContent.RT_MANIFEST.'#1'.'0409'.identity.version = $versionFull
                $winresContent | ConvertTo-Json -Depth 10 | Set-Content $winresFile
            }
        }
        go-winres make --in winres.json --arch amd64
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Windows resources generated (icon embedded)." -ForegroundColor Green
        } else {
            Write-Host "WARNING: Failed to generate Windows resources." -ForegroundColor Yellow
        }
    } else {
        Write-Host "go-winres not found. Install with: go install github.com/tc-hib/go-winres@latest" -ForegroundColor Yellow
        Write-Host "Continuing without embedded icon..." -ForegroundColor DarkGray
    }

    # Build ldflags with version
    $ldflags = "-X main.Version=$Version"

    if ($Dev) {
        Write-Host "`nStarting dev server..." -ForegroundColor Yellow
        wails3 dev
    }
    elseif ($Release) {
        Write-Host "`nBuilding release (single-exe with embedded DLL)..." -ForegroundColor Yellow
        
        # Add optimization flags for release
        # -s: strip symbol table, -w: strip DWARF debug info
        # -H=windowsgui: Windows GUI app (no console)
        $ldflags = "-s -w -H=windowsgui -X main.Version=$Version"
        
        # Build with Go directly (wails3 build has issues with ldflags)
        # -tags production: disable DevTools in release
        # -trimpath: remove file system paths from binary
        # DLL is embedded via go:embed in dll_embed.go
        $env:CGO_ENABLED = "1"
        go build -tags production -ldflags="$ldflags" -trimpath -o "build\bin\FKey.exe" .
        
        # Output directory
        $OutputDir = Join-Path $ProjectDir "build\bin"
        if (!(Test-Path $OutputDir)) {
            New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
        }
        
        # Remove old DLL from output dir if exists (single-exe now)
        $OldDll = Join-Path $OutputDir "gonhanh_core.dll"
        if (Test-Path $OldDll) {
            Remove-Item $OldDll -Force
        }
        
        # Check size
        $ExePath = Join-Path $OutputDir "FKey.exe"
        if (Test-Path $ExePath) {
            $Size = (Get-Item $ExePath).Length / 1MB
            Write-Host "`nBuild complete!" -ForegroundColor Green
            Write-Host "Version: $Version" -ForegroundColor Cyan
            Write-Host "Executable size: $([math]::Round($Size, 2)) MB" -ForegroundColor Cyan
            
            # Code signing
            Sign-Binary $ExePath
            
            Write-Host "`nOutput: $ExePath" -ForegroundColor Green
        }
    }
    else {
        Write-Host "`nBuilding debug..." -ForegroundColor Yellow
        go build -ldflags="$ldflags" -o FKey.exe .
        
        if (Test-Path "FKey.exe") {
            Write-Host "`nBuild complete!" -ForegroundColor Green
            Write-Host "Version: $Version" -ForegroundColor Cyan
            Write-Host "Run: .\FKey.exe" -ForegroundColor Cyan
        }
    }
}
finally {
    Pop-Location
}
