![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/bladeacer/mnemosync?style=for-the-badge&logo=go)
![GitHub License](https://img.shields.io/github/license/bladeacer/mnemosync?style=for-the-badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/bladeacer/mmsync)](https://goreportcard.com/report/github.com/bladeacer/mmsync)

# mns

Short for mnemosync. A CLI tool that lets you add folders to backup manually to
a target Git repository.

The name is inspired by the Greek Goddess of memory Mnemosyne.

## Installation

### Via Go (development)

```bash
go install github.com/bladeacer/mmsync@latest
```

### Via binary release

Download the latest binary for your platform from the
[releases page](https://github.com/bladeacer/mmsync/releases), extract it, and
place it in your `$PATH`.

**Always backup your files before using mmsync**.

```bash
mns
```

## Quick start

```bash
# Initialise with a target Git repository
mns init -r /path/to/backup-repo

# Add directories to track
mns add ~/Documents
mns add ~/Pictures --alias mypics

# Stage them (rsyncs into the repo's .mnemosync/staging/)
mns stage

# Archive, commit, and push
mns push
```

## Project status

Stable. See [this GitHub project](https://github.com/users/bladeacer/projects/3) for
the progress tracker.

## CLI reference

| Command | What it does | Status |
| --- | --- | --- |
| `mns` | Main help text to stdout | Done |
| `mns init` | Initialise a new configuration file with default values | Done |
| `mns add <path>... [-a alias]` | Add one or more directories to track for backup | Done |
| `mns list` | List all tracked directories | Done |
| `mns rm <id-or-alias>` | Remove a tracked directory | Done |
| `mns clear` | Remove all tracked directories (with confirmation) | Done |
| `mns search <query>` | Search tracked directories by path or alias | Done |
| `mns change <id-or-alias> [--path] [--alias]` | Change the path or alias of a tracked directory | Done |
| `mns stage [id-or-alias...]` | Rsync tracked directories to the repo staging area | Done |
| `mns unstage [id-or-alias...]` | Remove tracked directories from the staging area | Done |
| `mns status` | Show `git status` of the target repository | Done |
| `mns log [-n limit]` | Show `git log` of the target repository | Done |
| `mns push [--no-push]` | Archive staged files, commit, and push to remote | Done |
| `mns config` | Manage configuration file path | Done |
| `mns config get` | Print configuration file content | Done |
| `mns config open` | Open configuration file in `$EDITOR` | Done |
| `mns repo get` | Print the configured repository path | Done |
| `mns repo open` | Open the configuration file in `$EDITOR` | Done |
| `mns health` | Check required binaries (`git`, `rsync`, `tar`, `zip`) and config | Done |
| `mns version` | Print the version of mnemosync | Done |
| `mns man` | Generate and display the manual page | Done |
| `mns completion [shell]` | Generate autocompletion script (Cobra CLI builtin) | Done |
| `mns get-archiver` | Show the current archiver (`tar` or `zip`, default: `tar`) | Done |
| `mns set-archiver <tar|zip>` | Set the archiver | Done |
| `mns get-commit-fmt` | Show the commit message format (Go time layout) | Done |
| `mns set-commit-fmt <format>` | Set the commit message format | Done |
| `mns get-ignore` | Show whether `.gitignore` is respected during staging (`1`/`0`) | Done |
| `mns set-ignore <0|1>` | Set whether to respect `.gitignore` during staging | Done |
| `mns get-hist-limit` | Show history retention limits (days, max MB) | Done |
| `mns set-hist-limit -d <days> -s <size_mb>` | Set history retention limits | Done |
| `mns clear-hist` | Clear staging area and recorded history (with confirmation) | Done |

## Workflow

```
mns add ~/Documents              # register a directory
mns stage                        # rsync it into <repo>/.mnemosync/staging/
mns status                       # check git status in the repo
mns push                         # archive staging dir, commit, push
```

- Staging files are stored under `<repo>/.mnemosync/staging/` which is automatically
  gitignored — individual files are never tracked.
- On `push`, the staging directory is archived with `tar` (default) or `zip`,
  the archive is committed, and the remote is pushed.
- Only the 5 most recent archives are kept in the repo (configurable via
  `keep_archives`).
- Archives exceeding 5 MB trigger Git LFS tracking if `git-lfs` is installed
  (configurable via `lfs_threshold_mb`).

## Configuration

The configuration file is created at `~/.config/mmsync/config.yaml` (or
`$MMSYNC_CONF` if set) after running `mns init`. Default values:

| Field | Default | Description |
| --- | --- | --- |
| `archiver` | `tar` | Archive tool (`tar` or `zip`) |
| `commit_fmt` | `mnemosync archive 2006-01-02` | `time.Format` layout for commit messages |
| `respect_gitignore` | `true` | Whether to exclude `.gitignore` patterns during rsync |
| `hist_limit_days` | `7` | Days to retain staging history |
| `hist_limit_size_mb` | `1024` | Max size of staging history in MB |
| `keep_archives` | `5` | Number of recent archives to keep in the repo |
| `lfs_threshold_mb` | `5` | Archive size threshold to auto-configure Git LFS |

## Required dependencies

- `git` — version control
- `rsync` — staging mirroring (pre-installed on macOS and most Linux distros; on
  Windows use [MSYS2](https://www.msys2.org/) or [cwRsync](https://www.itefix.net/cwrsync))
- `tar` or `zip` — at least one archiver must be available (pre-installed on
  macOS and Linux; on Windows use [MSYS2](https://www.msys2.org/))
- `git-lfs` — optional, auto-configured for archives exceeding `lfs_threshold_mb`

## Development

```bash
git clone https://github.com/bladeacer/mmsync
cd mmsync
make build       # builds the mns binary
make lint        # run golangci-lint
make test        # run all tests
make gowatch     # hot-reload during active development
make snapshot    # test goreleaser locally (builds all platforms)
```

### Release

Releases are built with [GoReleaser](https://goreleaser.com):

```bash
make tag         # prompts for a version and creates an annotated tag
git push origin --tags
goreleaser release --clean
```

## LLM Usage Disclosure

As this is my first project using Golang, I initially used some AI assistance for
syntax, with the larger commits in the earlier iterations of the codebase.

These days, I have been trying to avoid relying on LLMs too much.

## License

This Golang CLI app, "mnemosync" is released under the GNU General Public
License version 3 (GPLv3) License.

### License Notice

```
This file is part of mnemosync. mnemosync is a CLI tool that lets you add
folders to backup manually to a target Git repository.

Copyright (c) 2025 bladeacer

mnemosync is free software: you can redistribute it and/or modify it under the
terms of the GNU General Public License as published by the Free Software
Foundation, either version 3 of the License, or (at your option) any later version.

mnemosync is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR
PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with mnemosync.
If not, see <https://www.gnu.org/licenses/>.
```

### License file

You can find the [license file here](./LICENSE).

## Credits

This CLI was made possible by [Cobra CLI](https://github.com/spf13/cobra).
