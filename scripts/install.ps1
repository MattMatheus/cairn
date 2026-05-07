param(
  [string]$Repo = $(if ($env:CAIRN_REPO) { $env:CAIRN_REPO } else { "MattMatheus/cairn" }),
  [string]$Version = $(if ($env:CAIRN_VERSION) { $env:CAIRN_VERSION } else { "latest" }),
  [string]$InstallDir = $(if ($env:CAIRN_INSTALL_DIR) { $env:CAIRN_INSTALL_DIR } else { Join-Path $HOME ".cairn\bin" })
)

$ErrorActionPreference = "Stop"

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
  "AMD64" { "amd64" }
  "ARM64" { "arm64" }
  default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

$asset = "cairn_windows_${arch}.zip"
if ($Version -eq "latest") {
  $url = "https://github.com/$Repo/releases/latest/download/$asset"
} else {
  $url = "https://github.com/$Repo/releases/download/$Version/$asset"
}

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
  $archive = Join-Path $tmp $asset
  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
  Write-Host "Downloading $url"
  Invoke-WebRequest -Uri $url -OutFile $archive
  Expand-Archive -Path $archive -DestinationPath $tmp -Force
  Copy-Item -Path (Join-Path $tmp "cairn.exe") -Destination (Join-Path $InstallDir "cairn.exe") -Force
  Write-Host "Installed cairn to $(Join-Path $InstallDir "cairn.exe")"
  & (Join-Path $InstallDir "cairn.exe") version
  if (($env:PATH -split ";") -notcontains $InstallDir) {
    Write-Host "Add $InstallDir to PATH before running cairn from a new shell."
  }
} finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
