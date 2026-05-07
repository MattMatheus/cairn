# Install

Pilots should use a released binary and a normal workspace directory. They should not need the Go toolchain or a repository-local `bin/` shim.

## macOS And Linux

```sh
curl -fsSL https://raw.githubusercontent.com/MattMatheus/cairn/main/scripts/install.sh | sh
export PATH="$HOME/.local/bin:$PATH"
cairn version
```

The installer downloads the current GitHub Release asset for your OS and architecture, then installs `cairn` to `~/.local/bin` by default. Release assets include ARM macOS.

Useful overrides:

```sh
CAIRN_VERSION=v0.1.0 sh scripts/install.sh
CAIRN_INSTALL_DIR=/usr/local/bin sh scripts/install.sh
CAIRN_REPO=MattMatheus/cairn sh scripts/install.sh
```

## Windows

From PowerShell:

```powershell
iwr https://raw.githubusercontent.com/MattMatheus/cairn/main/scripts/install.ps1 -UseB | iex
$env:PATH="$HOME\.cairn\bin;$env:PATH"
cairn version
```

The Windows installer downloads `cairn_windows_amd64.zip` or `cairn_windows_arm64.zip` and installs `cairn.exe` to `$HOME\.cairn\bin` by default.

Useful overrides:

```powershell
$env:CAIRN_VERSION="v0.1.0"
$env:CAIRN_INSTALL_DIR="$HOME\bin"
iwr https://raw.githubusercontent.com/MattMatheus/cairn/main/scripts/install.ps1 -UseB | iex
```

## Source Contributor Fallback

From a checkout:

```sh
go run ./cmd/cairn version
```

If you build locally, write the binary outside the repo or to an ignored contributor path, then use `--root` to point Cairn at a real workspace:

```sh
go build -o "$HOME/.local/bin/cairn" ./cmd/cairn
WORK_ROOT="$HOME/CairnPilot"
cairn --root "$WORK_ROOT" init
```

Cairn refuses `init` inside the Cairn source repository unless `--force` is provided.
