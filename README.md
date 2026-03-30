# tars

**tars** is package manager like homebrew which connects tools to multiple agents like Cursor, Claude Code, Gemini CLI, and Pi.

## How it fits together

**tars** contains two main modules:

1. **tars cli** (this repository): It is a package manager like homebrew which connects tools to multiple agents like Cursor, Claude Code, Gemini CLI, and Pi. It is used by users to install tools and connect them to agents.
2. **tars tools**: tars tools is like homebrew-core which contains the core tools and formulas. tools publishers can publish their tools to tars tools.

tars cli internally uses tars tools to install tools and connect them to agents.

## Install

To install tars cli go to Releases in github and download the latest release for your operating system.

After installation Unpack the release and run the following command to add tars to your PATH:

```bash
./tars link
```

if you get can’t be opened because Apple cannot check it for malicious software error, you can open it by running the following command:

```bash
xattr -d com.apple.quarantine tars
```

Then run the following command to add tars to your PATH:

```bash
./tars link
```
