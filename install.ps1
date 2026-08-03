[CmdletBinding()]
param(
    [string]$Version = $env:IMPRUN_VERSION,
    [string]$InstallDir = $env:IMPRUN_INSTALL_DIR,
    [switch]$NoModifyPath,
    [switch]$RequireSignature,
    [string]$ReleaseBaseUrl = $env:IMPRUN_RELEASE_BASE_URL
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repository = "imprun/cli"
if ([string]::IsNullOrWhiteSpace($ReleaseBaseUrl)) {
    $ReleaseBaseUrl = "https://github.com/$repository/releases/download"
}
if ([string]::IsNullOrWhiteSpace($InstallDir)) {
    if (-not [string]::IsNullOrWhiteSpace($env:LOCALAPPDATA)) {
        $InstallDir = Join-Path $env:LOCALAPPDATA "Programs\Imprun\bin"
    } else {
        $InstallDir = Join-Path $HOME ".local\bin"
    }
}
if ($env:IMPRUN_NO_MODIFY_PATH -match '^(1|true|yes)$') {
    $NoModifyPath = $true
}
if ($env:IMPRUN_REQUIRE_SIGNATURE -match '^(1|true|yes)$') {
    $RequireSignature = $true
}

function Receive-ReleaseFile {
    param(
        [Parameter(Mandatory = $true)][string]$Uri,
        [Parameter(Mandatory = $true)][string]$Destination
    )

    $parsedUri = [Uri]$Uri
    if ($parsedUri.IsFile) {
        Copy-Item -LiteralPath $parsedUri.LocalPath -Destination $Destination
        return
    }
    Invoke-WebRequest -UseBasicParsing -Uri $Uri -OutFile $Destination
}

function Get-Sha256 {
    param([Parameter(Mandatory = $true)][string]$Path)

    $stream = [IO.File]::OpenRead($Path)
    try {
        $sha256 = [Security.Cryptography.SHA256]::Create()
        try {
            return (($sha256.ComputeHash($stream) | ForEach-Object { $_.ToString('x2') }) -join '')
        } finally {
            $sha256.Dispose()
        }
    } finally {
        $stream.Dispose()
    }
}

if ([string]::IsNullOrWhiteSpace($Version)) {
    try {
        $latest = Invoke-RestMethod -UseBasicParsing -Uri "https://api.github.com/repos/$repository/releases/latest"
        $Version = [string]$latest.tag_name
    } catch {
        throw "Could not resolve the latest stable Imprun release: $($_.Exception.Message)"
    }
}

$versionMatch = [regex]::Match($Version, '^(?:v)?(?<version>[0-9]+\.[0-9]+\.[0-9]+(?:[+-][0-9A-Za-z][0-9A-Za-z.-]*)?)$')
if (-not $versionMatch.Success) {
    throw "Invalid Imprun version: $Version"
}
$Version = $versionMatch.Groups['version'].Value
$tag = "v$Version"

$runtimeArchitecture = $env:PROCESSOR_ARCHITEW6432
if ([string]::IsNullOrWhiteSpace($runtimeArchitecture)) {
    $runtimeArchitecture = $env:PROCESSOR_ARCHITECTURE
}
if ([string]::IsNullOrWhiteSpace($runtimeArchitecture)) {
    throw "Could not determine the Windows architecture"
}
switch ($runtimeArchitecture.ToUpperInvariant()) {
    "AMD64" { $architecture = "amd64" }
    "ARM64" { $architecture = "arm64" }
    default { throw "Unsupported Windows architecture: $runtimeArchitecture" }
}

$asset = "imprun_${Version}_windows_${architecture}.exe"
$releaseUrl = "$($ReleaseBaseUrl.TrimEnd('/'))/$tag"
$temporaryDir = Join-Path ([IO.Path]::GetTempPath()) "imprun-install-$([Guid]::NewGuid().ToString('N'))"
$stagedTarget = $null

try {
    New-Item -ItemType Directory -Path $temporaryDir | Out-Null
    $assetPath = Join-Path $temporaryDir $asset
    $checksumsPath = Join-Path $temporaryDir "checksums.txt"
    $bundlePath = Join-Path $temporaryDir "checksums.txt.sigstore.json"

    Receive-ReleaseFile -Uri "$releaseUrl/$asset" -Destination $assetPath
    Receive-ReleaseFile -Uri "$releaseUrl/checksums.txt" -Destination $checksumsPath

    $cosign = Get-Command cosign -ErrorAction SilentlyContinue
    if ($null -ne $cosign) {
        Receive-ReleaseFile -Uri "$releaseUrl/checksums.txt.sigstore.json" -Destination $bundlePath
        & $cosign.Source verify-blob `
            --bundle $bundlePath `
            --certificate-identity "https://github.com/$repository/.github/workflows/release.yml@refs/tags/$tag" `
            --certificate-oidc-issuer "https://token.actions.githubusercontent.com" `
            $checksumsPath | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw "Release signature verification failed"
        }
    } elseif ($RequireSignature) {
        throw "cosign is required when signature verification is mandatory"
    } else {
        Write-Warning "cosign was not found; SHA-256 was verified but signer verification was skipped"
    }

    $escapedAsset = [regex]::Escape($asset)
    $checksumEntries = @(
        Get-Content -LiteralPath $checksumsPath | ForEach-Object {
            $match = [regex]::Match($_, "^(?<hash>[0-9A-Fa-f]{64})\s+\*?$escapedAsset$")
            if ($match.Success) { $match.Groups['hash'].Value.ToLowerInvariant() }
        }
    )
    if ($checksumEntries.Count -ne 1) {
        throw "The release checksum must contain exactly one entry for $asset"
    }
    $actualChecksum = Get-Sha256 -Path $assetPath
    if ($actualChecksum -ne $checksumEntries[0]) {
        throw "SHA-256 mismatch for $asset"
    }

    $candidateVersion = @(& $assetPath --version 2>&1)
    if ($LASTEXITCODE -ne 0) {
        throw "Downloaded executable did not run"
    }
    $reportedVersion = ($candidateVersion -join "`n").Trim()
    if ($reportedVersion -ne $Version -and $reportedVersion -ne "imprun $Version") {
        throw "Downloaded executable reported an unexpected version: $reportedVersion"
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    $target = Join-Path $InstallDir "imprun.exe"
    $stagedTarget = Join-Path $InstallDir ".imprun.new-$PID.exe"
    Copy-Item -LiteralPath $assetPath -Destination $stagedTarget
    & $stagedTarget --version | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Staged executable did not run"
    }
    Move-Item -LiteralPath $stagedTarget -Destination $target -Force
    $stagedTarget = $null

    if (-not $NoModifyPath) {
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        $userEntries = @($userPath -split ';' | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
        if (-not ($userEntries | Where-Object { [string]::Equals($_.TrimEnd('\'), $InstallDir.TrimEnd('\'), [StringComparison]::OrdinalIgnoreCase) })) {
            $updatedUserPath = (@($userEntries) + $InstallDir) -join ';'
            [Environment]::SetEnvironmentVariable("Path", $updatedUserPath, "User")
        }
        $processEntries = @($env:Path -split ';')
        if (-not ($processEntries | Where-Object { [string]::Equals($_.TrimEnd('\'), $InstallDir.TrimEnd('\'), [StringComparison]::OrdinalIgnoreCase) })) {
            $env:Path = "$InstallDir;$env:Path"
        }
    }

    Write-Output "Installed imprun $Version to $target"
    if ($NoModifyPath) {
        Write-Output "Add $InstallDir to PATH, then run: imprun --version"
    } else {
        Write-Output "Open a new terminal, then run: imprun --version"
    }
} finally {
    if ($null -ne $stagedTarget -and (Test-Path -LiteralPath $stagedTarget)) {
        Remove-Item -LiteralPath $stagedTarget -Force
    }
    if (Test-Path -LiteralPath $temporaryDir) {
        Remove-Item -LiteralPath $temporaryDir -Recurse -Force
    }
}
