#!/usr/bin/env sh
set -eu

repo="${CAIRN_REPO:-MattMatheus/cairn}"
version="${CAIRN_VERSION:-latest}"
install_dir="${CAIRN_INSTALL_DIR:-$HOME/.local/bin}"

case "$repo" in
  */*/*|/*|*"://"*|"" )
    echo "Ignoring CAIRN_REPO=$repo; expected GitHub owner/repo such as MattMatheus/cairn." >&2
    repo="MattMatheus/cairn"
    ;;
  */*) ;;
  *)
    echo "Ignoring CAIRN_REPO=$repo; expected GitHub owner/repo such as MattMatheus/cairn." >&2
    repo="MattMatheus/cairn"
    ;;
esac

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
esac

case "$os" in
  darwin|linux) ;;
  *) echo "unsupported operating system: $os" >&2; exit 1 ;;
esac

asset="cairn_${os}_${arch}.tar.gz"

if [ "$version" = "latest" ]; then
  latest_json="$(curl -fsSL "https://api.github.com/repos/${repo}/releases/latest" 2>/dev/null || true)"
  version="$(printf '%s' "$latest_json" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
  if [ -z "$version" ]; then
    releases_json="$(curl -fsSL "https://api.github.com/repos/${repo}/releases?per_page=20" 2>/dev/null || true)"
    version="$(printf '%s' "$releases_json" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
  fi
  if [ -z "$version" ]; then
    echo "could not find a published Cairn release; set CAIRN_VERSION to a tag such as v0.N" >&2
    exit 1
  fi
  echo "Resolved latest Cairn release: $version"
else
  echo "Using requested Cairn release: $version"
fi

url="https://github.com/${repo}/releases/download/${version}/${asset}"
checksum_url="https://github.com/${repo}/releases/download/${version}/cairn_${os}_${arch}.sha256"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

mkdir -p "$install_dir"
echo "Downloading $url"
curl -fsSL "$url" -o "$tmp_dir/$asset"
if ! curl -fsSL "$checksum_url" -o "$tmp_dir/$asset.sha256"; then
  echo "failed to download checksum from $checksum_url; refusing to install unverified binary" >&2
  exit 1
fi
if command -v sha256sum >/dev/null 2>&1; then
  (cd "$tmp_dir" && sha256sum -c "$asset.sha256")
else
  expected="$(awk '{print $1}' "$tmp_dir/$asset.sha256")"
  actual="$(shasum -a 256 "$tmp_dir/$asset" | awk '{print $1}')"
  if [ "$actual" != "$expected" ]; then
    echo "checksum mismatch for $asset" >&2
    exit 1
  fi
  echo "$asset: OK"
fi
tar -xzf "$tmp_dir/$asset" -C "$tmp_dir"
install "$tmp_dir/cairn" "$install_dir/cairn"

echo "Installed cairn to $install_dir/cairn"
installed_version="$("$install_dir/cairn" version)"
echo "$installed_version"
case "$installed_version" in
  *"$version"*) ;;
  *) echo "Warning: installed binary version did not include expected release $version" >&2 ;;
esac
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "Add $install_dir to PATH before running cairn from a new shell." ;;
esac
