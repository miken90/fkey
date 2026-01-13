# Simple Build Script
param([string]$Configuration = "Debug")

$RootDir = $PSScriptRoot

Write-Host "Building GoNhanh..." -ForegroundColor Cyan

# Step 1: Build Rust core DLL
Write-Host "Step 1: Building Rust core..." -ForegroundColor Yellow
Push-Location "$RootDir/core"
try {
    if ($Configuration -eq "Release") {
        cargo build --release
    } else {
        cargo build --release  # Always use release for DLL (smaller, faster)
    }
    if ($LASTEXITCODE -ne 0) {
        throw "Rust build failed"
    }
}
finally {
    Pop-Location
}

# Step 2: Copy DLL to Native folder
Write-Host "Step 2: Copying DLL..." -ForegroundColor Yellow
$DllSource = "$RootDir/core/target/release/gonhanh_core.dll"
$DllDest = "$RootDir/platforms/windows/GoNhanh/Native/gonhanh_core.dll"
Copy-Item -Path $DllSource -Destination $DllDest -Force
Write-Host "  Copied: $DllSource -> $DllDest" -ForegroundColor Gray

# Step 3: Build Windows app
Write-Host "Step 3: Building Windows app..." -ForegroundColor Yellow
Push-Location "$RootDir/platforms/windows/GoNhanh"
try {
    dotnet clean --configuration $Configuration --nologo
    dotnet build --configuration $Configuration --nologo
    if ($LASTEXITCODE -ne 0) {
        throw "dotnet build failed"
    }
    Write-Host "Build completed!" -ForegroundColor Green
}
catch {
    Write-Host "Build failed: $_" -ForegroundColor Red
    exit 1
}
finally {
    Pop-Location
}
