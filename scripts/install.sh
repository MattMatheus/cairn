#!/usr/bin/env sh
set -eu

repo="${CAIRN_REPO:-MattMatheus/cairn}"
version="${CAIRN_VERSION:-latest}"
install_dir="${CAIRN_INSTALL_DIR:-$HOME/.local/bin}"

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
  url="https://github.com/${repo}/releases/latest/download/${asset}"
else
  url="https://github.com/${repo}/releases/download/${version}/${asset}"
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

mkdir -p "$install_dir"
echo "Downloading $url"
curl -fsSL "$url" -o "$tmp_dir/$asset"
tar -xzf "$tmp_dir/$asset" -C "$tmp_dir"
install "$tmp_dir/cairn" "$install_dir/cairn"

echo "Installed cairn to $install_dir/cairn"
if command -v "$install_dir/cairn" >/dev/null 2>&1; then
  "$install_dir/cairn" version
fi
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "Add $install_dir to PATH before running cairn from a new shell." ;;
esac
