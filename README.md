# tars

**tars** is package manager like homebrew which connects tools to multiple agents like Cursor, Claude Code, Gemini CLI, and Pi.

## How does it work?

**tars** contains two main modules:

1. **tars cli** (this repository): It is a package manager like homebrew which connects tools to multiple agents like Cursor, Claude Code, Gemini CLI, and Pi. It is used by users to install tools and connect them to agents.
2. **tars tools**: tars tools is like homebrew-core which contains the core tools and formulas. tools publishers can publish their tools to tars tools.

tars cli internally uses tars tools to install tools and connect them to agents.

## Install

1. To install tars cli go to [Releases](https://github.com/chenchunaidu/tars/releases/) in github and download the latest release for your operating system.

2. After installation Unpack the release and run the following command to add tars to your PATH:

```bash
./tars link
```

3. if you get can’t be opened because Apple cannot check it for malicious software error, you can open it by running the following command:

```bash
xattr -d com.apple.quarantine tars
```

4. Then run the following command to add tars to your PATH:

```bash
./tars link
```

## Commands

Global flags: `-h` / `--help`, `-v` / `--version`. Use `tars help <command>` or `tars <command> --help` for full usage.

| Command | Usage | Description |
| --- | --- | --- |
| **link** | `tars link` | Symlink the `tars` binary to `~/.tars/bin` and add that directory to your shell `PATH` (or Windows user `PATH`). Open a new terminal afterward. |
| **update** | `tars update` | Pull the latest formula definitions for the core tap and any extra taps. |
| **install** | `tars install <formula name or path>` | Download the artifact, verify SHA256, install under `~/.tars`, refresh `tools.json` / `tools.md`, and connect agents. |
| **uninstall** | `tars uninstall <name>` | Remove an installed tool, update the catalog and agent wiring. |
| **list** | `tars list` | List installed tools (name, version, tap). |
| | `tars list --available` / `-a` | List formula names available from taps. |
| **info** | `tars info <name>` | Show details for an installed tool, or pretty-print the formula JSON if not installed. |
| **catalog** | `tars catalog` | Print the path to the merged model catalog (`~/.tars/catalog/tools.json`). |
| **hash** | `tars hash <file>` | Print the file’s SHA256 (for filling formula `sha256` fields). |
| **connect** | `tars connect all` | Regenerate `~/.tars/tools.md` and update global instructions for **all** agents (Cursor, Claude Code, Gemini CLI, Pi). |
| | `tars connect <agent> [agent...]` | Same, but only for named agents: `cursor`, `claude`, `gemini`, `pi`. |
| | `tars connect ... --copy <dir>` | Also copy `tools.md` into `<dir>/tools.md` (e.g. `--copy .` for the current project). |
| **tap** | `tars tap add <name> <git-url>` | Clone an extra formula tap (core is implicit; override core with `TARS_CORE_URL`). |
| | `tars tap list` | List core and registered taps. |
| **publish** | `tars publish validate <formula.json>` | Validate formula JSON and required security fields. |
| | `tars publish init <name>` | Write a template `<name>.json` in the current directory. |
| **completion** | `tars completion bash` (or `fish`, `powershell`, `zsh`) | Emit shell completion scripts; see `tars completion <shell> --help` for how to load them. |

**Environment:** set `TARS_CORE_URL` to use a non-default core formula repository URL.
