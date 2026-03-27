# GitHub Packages (npm)

Releases are triggered only by pushing a **`v*`** git tag. Each release publishes **three** scoped npm packages to **GitHub Packages**—one per **platform** (OS). Each package bundles **amd64 and arm64** binaries; the `tars` command picks the right one at runtime (via the small Node launcher in `bin/tars.js`).

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
