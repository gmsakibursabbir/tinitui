# TiniTUI Windows Installer
$ErrorActionPreference = "Stop"

$RepoOwner = "gmsakibursabbir"
$RepoName = "tinitui"
$BinaryName = "tinitui.exe"

Write-Host "Installing TiniTUI..." -ForegroundColor Cyan

# 1. Detect Architecture
$Arch = $env:PROCESSOR_ARCHITECTURE
if ($Arch -eq "AMD64") {
    $AssetArch = "amd64"
} elseif ($Arch -eq "ARM64") {
    $AssetArch = "arm64"
} else {
    Write-Error "Unsupported architecture: $Arch"
    exit 1
}

# 2. Find Latest Release
$LatestUrl = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
try {
    $Release = Invoke-RestMethod -Uri $LatestUrl
    $TagName = $Release.tag_name
} catch {
    Write-Error "Failed to fetch latest release version. check your internet connection."
    exit 1
}

Write-Host "Latest Version: $TagName" -ForegroundColor Green

# 3. Construct Download URL
$AssetName = "tinytui-windows-$AssetArch.exe"
$DownloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$TagName/$AssetName"

# 4. Determine Install Directory
# Try to find a directory in PATH that is user-writable
$InstallDir = "$env:LOCALAPPDATA\Programs\TiniTUI"
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

# Add to PATH if not present
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "Adding $InstallDir to User PATH..." -ForegroundColor Yellow
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    $env:Path += ";$InstallDir"
}

# 5. Download and Install
$OutputFile = Join-Path $InstallDir $BinaryName
Write-Host "Downloading from $DownloadUrl..."
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $OutputFile
} catch {
    Write-Error "Failed to download binary. Please try again."
    exit 1
}

Write-Host "Success! Installed to $OutputFile" -ForegroundColor Green
Write-Host "You may need to restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
Write-Host "Run 'tinitui' to start." -ForegroundColor Cyan
