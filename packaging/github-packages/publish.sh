#!/usr/bin/env bash
# Assemble three platform npm packages (linux, darwin, windows) from release build artifacts
# and publish to GitHub Packages. Each package ships amd64 + arm64 binaries.
#
# Required env: VERSION (e.g. v1.2.3), SEMVER (1.2.3), OWNER_LC (lowercase GitHub owner),
#               GITHUB_REPOSITORY (owner/repo), NODE_AUTH_TOKEN, ARTIFACTS (path to merged artifacts).

set -euo pipefail

ROOT=$(cd "$(dirname "$0")" && pwd)
ARTIFACTS=${ARTIFACTS:-artifacts}
NPM_DIST=${NPM_DIST:-npm-packages-out}

: "${VERSION:?}"
: "${SEMVER:?}"
: "${OWNER_LC:?}"
: "${GITHUB_REPOSITORY:?}"
: "${NODE_AUTH_TOKEN:?}"

rm -rf "$NPM_DIST"
mkdir -p "$NPM_DIST"

write_npmrc() {
  cat > "${HOME}/.npmrc" <<EOF
//npm.pkg.github.com/:_authToken=${NODE_AUTH_TOKEN}
@${OWNER_LC}:registry=https://npm.pkg.github.com
EOF
}

copy_launcher() {
  local dest=$1
  mkdir -p "$dest/bin"
  cp "$ROOT/tars-bin.js" "$dest/bin/tars.js"
  chmod +x "$dest/bin/tars.js"
}

make_pkg() {
  local platform=$1
  local goos=$2
  local pkg="$NPM_DIST/tars-$platform"
  mkdir -p "$pkg/libexec"
  copy_launcher "$pkg"

  if [ "$goos" = "windows" ]; then
    local amd arch_zip d_amd d_arm
    amd=$(find "$ARTIFACTS" -name "tars_${VERSION}_windows_amd64.zip" -print -quit)
    arch_zip=$(find "$ARTIFACTS" -name "tars_${VERSION}_windows_arm64.zip" -print -quit)
    [ -n "$amd" ] && [ -f "$amd" ] || {
      echo "missing windows amd64 zip under $ARTIFACTS" >&2
      exit 1
    }
    [ -n "$arch_zip" ] && [ -f "$arch_zip" ] || {
      echo "missing windows arm64 zip under $ARTIFACTS" >&2
      exit 1
    }
    d_amd=$(mktemp -d)
    d_arm=$(mktemp -d)
    unzip -j -o "$amd" -d "$d_amd"
    unzip -j -o "$arch_zip" -d "$d_arm"
    mv "$d_amd/tars.exe" "$pkg/libexec/tars-amd64.exe"
    mv "$d_arm/tars.exe" "$pkg/libexec/tars-arm64.exe"
    rm -rf "$d_amd" "$d_arm"
  else
    local amd_tgz arm_tgz d_amd d_arm
    amd_tgz=$(find "$ARTIFACTS" -name "tars_${VERSION}_${goos}_amd64.tar.gz" -print -quit)
    arm_tgz=$(find "$ARTIFACTS" -name "tars_${VERSION}_${goos}_arm64.tar.gz" -print -quit)
    [ -n "$amd_tgz" ] && [ -f "$amd_tgz" ] || {
      echo "missing ${goos} amd64 tarball under $ARTIFACTS" >&2
      exit 1
    }
    [ -n "$arm_tgz" ] && [ -f "$arm_tgz" ] || {
      echo "missing ${goos} arm64 tarball under $ARTIFACTS" >&2
      exit 1
    }
    d_amd=$(mktemp -d)
    d_arm=$(mktemp -d)
    tar -xzf "$amd_tgz" -C "$d_amd"
    tar -xzf "$arm_tgz" -C "$d_arm"
    mv "$d_amd/tars" "$pkg/libexec/tars-amd64"
    mv "$d_arm/tars" "$pkg/libexec/tars-arm64"
    chmod +x "$pkg/libexec/tars-amd64" "$pkg/libexec/tars-arm64"
    rm -rf "$d_amd" "$d_arm"
  fi

  cat >"$pkg/package.json" <<EOF
{
  "name": "@${OWNER_LC}/tars-${platform}",
  "version": "${SEMVER}",
  "description": "tars CLI for ${platform} (amd64 and arm64)",
  "bin": {
    "tars": "bin/tars.js"
  },
  "files": [
    "bin",
    "libexec"
  ],
  "publishConfig": {
    "registry": "https://npm.pkg.github.com"
  },
  "repository": {
    "type": "git",
    "url": "git+https://github.com/${GITHUB_REPOSITORY}.git"
  }
}
EOF
}

write_npmrc
make_pkg linux linux
make_pkg darwin darwin
make_pkg windows windows

export NPM_CONFIG_PROVENANCE=false

for p in linux darwin windows; do
  (cd "$NPM_DIST/tars-$p" && npm publish --access public)
done
