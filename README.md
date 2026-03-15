![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/bladeacer/mnemosync?style=for-the-badge&logo=go)
![GitHub License](https://img.shields.io/github/license/bladeacer/mnemosync?style=for-the-badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/bladeacer/mmsync)](https://goreportcard.com/report/github.com/bladeacer/mmsync)

# mns

Short for mnemosync. A CLI tool that lets you add folders to backup manually to
a target Git repository.

The name is inspired by the Greek Goddess of memory Mnemosyne.

## Installation guide

This installation guide assumes you know how to create and set up a Git repository.

**Always backup your files before using mmsync**.

```bash
go install github.com/bladeacer/mmsync@latest
```

Ensure that you can access Go binaries in your $PATH.

```bash
mns
```

## Project status

WIP. See [this GitHub project](https://github.com/users/bladeacer/projects/3) for
the progress tracker.

This is my first project using the Go programming language, but I hope it will
be useful.

## Planned features

- Check if required binaries are available before calling the tool
  - Required binaries: `git, rsync, tar, zip`

- Help command line flag

- CRUD target directories which user wishes to backup e.g.

- `rsync` to mirror said target directories to a `~/.mnemosync/folders`
  - Either manually triggered or we integrate `cron`
- Wrapper for user to manually copy the files and push them in their Git repository

- Wrapper to let user set default commit message format

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

## Planned CLI spec

View currently available options by running mnemosync without any flags or arguments.

There is also `mns man` for a generated manual page.

| Command | What it does | Implementation Status |
| --- | --- | --- |
| `mns` | Main help text to stdout. | Done |
| `mns init` | Initialise a new configuration file with default values. | Done |
| `mns config` | Manage configuration file. | WIP |
| `mns completion` | Generate autocompletion script for target shell (Cobra CLI builtin)  | Done |
| `mns health` | Check Health and installation of dependencies (optional or not) | Done |
| `mns help` | Get help text for specific command (Status depends on target command) | WIP |
| `mns man` | Generates the manual page and displays with less | WIP |
| `mns add` | Add target file/folder with target with alias support. | WIP |

<!-- ```bash -->
<!-- ## Init and config -->
<!-- # Init the app with helpers to get the user to set path config and all -->
<!-- mmsync init --> 
<!-- mmsync config -->
<!-- mmsync config open -->
<!-- # Prints to stdout -->
<!-- mmsync repo get -->

<!-- ## CRUD directories to mmsync before staging -->
<!-- # Save this in the local viewable db somehow each time the binary is called. -->
<!-- mmsync add <target_path> -a <optional_alias> -->
<!-- mmsync list -->
<!-- mmsync change <target_path-or-alias> <new-target_path-or-alias> -->
<!-- mmsync rm <target_path-or-alias> -->

<!-- ## Find a mmsync path or alias that has been added -->
<!-- mmsync search <query-by-path-or-alias> -->

<!-- ## Add warning for user to confirm if they wish to delete all directories they added -->
<!-- mmsync clear -->

<!-- # Backup related -->
<!-- ## Technical info: staging is just rsyncing over to the target repo -->
<!-- ## You can use . to include all directories and aliases -->

<!-- # rsyncs all added target mmsync directories or aliases to staging, and then -->
<!-- # calls git add in the target repo -->
<!-- mmsync stage <target_path-or-alias> --> 

<!-- # rsyncs unstages added target mmsync directories or aliases to staging --> 
<!-- # git restore --staged <target_path-or-alias> in the target repo -->
<!-- # somehow map aliases to directory names -->
<!-- mmsync unstage <target_path-or-alias> --> 

<!-- # get status of staging -->
<!-- # git status in the target repo -->
<!-- mmsync status -->

<!-- # get staging history --> 
<!-- # git log --oneline target repo -->
<!-- mmsync log -->

<!-- ## get staging history limit in days before it is cleared. Defaults to 7 days -->
<!-- ## and a max of 1024 MB. Limitation only enforced when the binary is called -->

<!-- ## Need to read last modified time of each rsync mirrored directory or save its -->
<!-- ## last modified time each time an operation is done on it. -->
<!-- ## with confirmation message -->

<!-- mmsync get-hist-limit --> 
<!-- mmsync set-hist-limit -d <number_of_days> -s <max_size_in_mb> -->

<!-- # calls git restore --staged . and git restore . in the target repo -->
<!-- # with confirmation message -->
<!-- mmsync clear-hist # clears staging history -->

<!-- # set archive options -->
<!-- # write this option in the config file somehow -->
<!-- mmsync set-archiver tar|zip -->
<!-- mmsync get-archiver # gets archive tool used, defaults to tar -->

<!-- # Git related -->

<!-- ## Configure commit messages -->
<!-- mmsync get-commit-fmt # Defaults to mnemosync archive ISO timestamp -->
<!-- mmsync set-commit-fmt <custom_format> -->

<!-- ## checks if anything in staging, if yes it compresses writes the archive file -->
<!-- ## over to be pushed -->
<!-- ## if not, warns the user that staging is empty or files are not staged yet -->
<!-- ## When pushing, write folder and filenames affected to viewable local db as -->
<!-- ## part of staging history -->
<!-- ## Also does the needed git commit and push on behalf of the user. -->
<!-- mmsync push --> 

<!-- ## Respecting .gitignore -->
<!-- ## mmsync respects gitignore in the target repo when adding directories or aliases -->
<!-- ## Returns true or 1 by default -->
<!-- ## When setting to 0, warning + confirmation -->
<!-- mmsync get-ignore --> 
<!-- mmsync set-ignore 0|1 -->

<!-- # Misc -->
<!-- mmsync version -->
<!-- mmsync help -->
<!-- ``` -->
