param(
    [string]$Version = "latest",
    [string]$InstallDir = (Join-Path $HOME ".local\bin")
)

$ErrorActionPreference = "Stop"
$repository = "scc-digitalhub/digitalhub-cli"

switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
    "X64" { $architecture = "amd64" }
    "Arm64" { $architecture = "arm64" }
    default { throw "Unsupported architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" }
}

if ($Version -eq "latest") {
    $releaseUrl = "https://github.com/$repository/releases/latest/download"
} else {
    $tag = if ($Version.StartsWith("v")) { $Version } else { "v$Version" }
    $releaseUrl = "https://github.com/$repository/releases/download/$tag"
}

$archive = "dhcli-windows-$architecture.zip"
$temporaryDirectory = Join-Path ([System.IO.Path]::GetTempPath()) ([guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $temporaryDirectory | Out-Null

try {
    $archivePath = Join-Path $temporaryDirectory $archive
    $checksumsPath = Join-Path $temporaryDirectory "checksums.txt"
    Invoke-WebRequest -Uri "$releaseUrl/$archive" -OutFile $archivePath
    Invoke-WebRequest -Uri "$releaseUrl/checksums.txt" -OutFile $checksumsPath

    $checksumLine = Get-Content $checksumsPath | Where-Object { $_ -match "\s+$([regex]::Escape($archive))$" } | Select-Object -First 1
    if (-not $checksumLine -or $checksumLine -notmatch "^([a-fA-F0-9]{64})\s+") {
        throw "No valid checksum found for $archive"
    }

    $expectedHash = $Matches[1].ToLowerInvariant()
    $actualHash = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actualHash -ne $expectedHash) {
        throw "Checksum verification failed for $archive"
    }

    $extractedDirectory = Join-Path $temporaryDirectory "extracted"
    Expand-Archive -Path $archivePath -DestinationPath $extractedDirectory
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item -Path (Join-Path $extractedDirectory "dhcli.exe") -Destination (Join-Path $InstallDir "dhcli.exe") -Force

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $pathEntries = @($userPath -split ";" | Where-Object { $_ })
    if (-not ($pathEntries | Where-Object { $_.TrimEnd("\") -ieq $InstallDir.TrimEnd("\") })) {
        $updatedPath = (@($pathEntries) + $InstallDir) -join ";"
        [Environment]::SetEnvironmentVariable("Path", $updatedPath, "User")
    }

    if (-not (($env:Path -split ";") | Where-Object { $_.TrimEnd("\") -ieq $InstallDir.TrimEnd("\") })) {
        $env:Path = "$InstallDir;$env:Path"
    }

    Write-Output "Installed dhcli to $(Join-Path $InstallDir 'dhcli.exe')"
    Write-Output "Open a new terminal if dhcli is not available in this one."
} finally {
    Remove-Item -Path $temporaryDirectory -Recurse -Force
}