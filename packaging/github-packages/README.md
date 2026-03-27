# GitHub Packages (npm)

This repo does **not** publish npm packages from CI. To publish **three** scoped packages to **GitHub Packages** yourself (one per OS, each with amd64 + arm64), download the release build artifacts, then run `publish.sh` with the env vars documented in that script. Each package uses the Node launcher in `bin/tars.js` to pick the right native binary.

| Package | Platform |
|---------|----------|
| `@OWNER/tars-linux` | Linux (x86_64 and arm64) |
| `@OWNER/tars-darwin` | macOS (Intel and Apple Silicon) |
| `@OWNER/tars-windows` | Windows (x86_64 and arm64) |

Replace `OWNER` with your GitHub user or organization (**lowercase**). Install the package that matches your operating system.

## Install from GitHub Packages

```bash
echo "@OWNER:registry=https://npm.pkg.github.com" >> ~/.npmrc
npm install -g @OWNER/tars-darwin@1.2.3
```

Use the **semver** from the tag (no `v` prefix). For private repos, authenticate with a token that has `read:packages`. You need **Node.js** on your machine for the global `tars` shim (the published packages delegate to the native binary).

**Binaries are also on [GitHub Releases](https://github.com) for this repository** (separate zip/tar per OS and architecture).
