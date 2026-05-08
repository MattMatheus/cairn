param(
  [string]$Repo = $(if ($env:CAIRN_REPO) { $env:CAIRN_REPO } else { "MattMatheus/cairn" }),
  [string]$Version = $(if ($env:CAIRN_VERSION) { $env:CAIRN_VERSION } else { "latest" }),
  [string]$InstallDir = $(if ($env:CAIRN_INSTALL_DIR) { $env:CAIRN_INSTALL_DIR } else { Join-Path $HOME ".cairn\bin" })
)

$ErrorActionPreference = "Stop"

if ($Repo -notmatch '^[^/\\:]+/[^/\\:]+$') {
  Write-Warning "Ignoring CAIRN_REPO=$Repo; expected GitHub owner/repo such as MattMatheus/cairn."
  $Repo = "MattMatheus/cairn"
}

$processorArch = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
$arch = switch ($processorArch) {
  "AMD64" { "amd64" }
  "ARM64" { "arm64" }
  default { throw "Unsupported architecture: $processorArch" }
}

$asset = "cairn_windows_${arch}.zip"
if ($Version -eq "latest") {
  try {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $release.tag_name
  } catch {
    $releases = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases?per_page=20"
    $release = $releases | Select-Object -First 1
    if (-not $release -or -not $release.tag_name) {
      throw "Could not find a published Cairn release. Set CAIRN_VERSION to a tag such as v0.N."
    }
    $Version = $release.tag_name
  }
  Write-Host "Resolved latest Cairn release: $Version"
} else {
  Write-Host "Using requested Cairn release: $Version"
}
$url = "https://github.com/$Repo/releases/download/$Version/$asset"
$checksumUrl = "https://github.com/$Repo/releases/download/$Version/cairn_windows_${arch}.sha256"

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
  $archive = Join-Path $tmp $asset
  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
  Write-Host "Downloading $url"
  Invoke-WebRequest -Uri $url -OutFile $archive
  try {
    $checksumPath = Join-Path $tmp "$asset.sha256"
    Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumPath
    $expected = ((Get-Content $checksumPath -Raw) -split "\s+")[0]
    $actual = (Get-FileHash -Algorithm SHA256 $archive).Hash.ToLowerInvariant()
    if ($actual -ne $expected.ToLowerInvariant()) {
      throw "Checksum mismatch for $asset"
    }
    Write-Host "$asset: OK"
  } catch {
    throw "Failed to verify $asset`: $($_.Exception.Message)"
  }
  Expand-Archive -Path $archive -DestinationPath $tmp -Force
  Copy-Item -Path (Join-Path $tmp "cairn.exe") -Destination (Join-Path $InstallDir "cairn.exe") -Force
  Write-Host "Installed cairn to $(Join-Path $InstallDir "cairn.exe")"
  $installedVersion = & (Join-Path $InstallDir "cairn.exe") version
  Write-Host $installedVersion
  if ($installedVersion -notlike "*$Version*") {
    Write-Warning "Installed binary version did not include expected release $Version"
  }
  if (($env:PATH -split ";") -notcontains $InstallDir) {
    Write-Host "Add $InstallDir to PATH before running cairn from a new shell."
  }
} finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
