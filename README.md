# CLI

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/scc-digitalhub/digitalhub-cli/release.yaml?event=release) [![license](https://img.shields.io/badge/license-Apache%202.0-blue)](https://github.com/scc-digitalhub/digitalhub-cli/LICENSE) ![GitHub Release](https://img.shields.io/github/v/release/scc-digitalhub/digitalhub-cli)
![Status](https://img.shields.io/badge/status-stable-gold)

The Command-Line Interface (CLI) is a simple tool that offers a number of functionalities to set the platform up.

The executable can be downloaded in the _Releases_ page.

For a full overview of available commands, please check [the documentation](https://scc-digitalhub.github.io/docs/0.11/cli/commands/).



## Quick start

Download and unpack the corresponding archive from the [GitHub release page](https://github.com/scc-digitalhub/digitalhub-cli/releases). Run the `dhcli` executable.

To install the latest release directly, use the installer for your platform. The installers verify the downloaded archive against the SHA-256 checksum published with the release and install it in your user directory without administrator privileges.

Linux and macOS (requires `curl` or `wget`, `tar`, and `sha256sum` or `shasum`):

```sh
curl -fsSL https://raw.githubusercontent.com/scc-digitalhub/digitalhub-cli/main/install.sh | sh
```

To install a specific version, pass it without the `v` prefix:

```sh
curl -fsSL https://raw.githubusercontent.com/scc-digitalhub/digitalhub-cli/main/install.sh | sh -s 0.15.2
```

Run the same command again to update. If `dhcli` is already installed, the script asks before replacing it (the default answer is yes); without a terminal, for example in CI, it replaces it without asking.

The binary is installed to `~/.local/bin`; set `DHCLI_INSTALL_DIR` to choose another directory. If that directory is not already on your `PATH`, add it to your shell configuration.

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/scc-digitalhub/digitalhub-cli/main/install.ps1 | iex
```

PowerShell installs to `$HOME\.local\bin` and adds that directory to your user `PATH`. To install a specific version or choose another directory, download `install.ps1` and run it with `-Version 0.15.2` and/or `-InstallDir <directory>`.

On Mac, you can also use the _Homebrew Tap_ distribution:

```sh
brew tap scc-digitalhub/digitalhub-cli https://github.com/scc-digitalhub/digitalhub-cli
brew install dhcli
```


## Commands

Run `dhcli --help` to see the available commands, or `dhcli <command> --help` for
command-specific usage and flags. The commands below operate on the selected
environment unless noted otherwise.

| Command | Description |
| --- | --- |
| `config` | Show the selected environment's non-secret configuration. Use the provider filter to show only one provider's settings. |
| `credentials` | Show credentials for the selected environment, optionally filtered by provider. **Output can contain secrets; handle it securely.** |
| `create <resource>` | Create a resource from a YAML definition. Projects can instead be created with a name; use `--reset-id` when the server must assign a new resource ID. |
| `delete <resource> [<id>]` | Delete a resource by ID, or use `--name` to target all versions by name. Confirmation is requested unless skipped; project deletion can optionally cascade to related resources. |
| `download <resource> [<id>]` | Download a resource's S3 content to a file or directory. Select by ID or name, and choose the destination and optional output format. |
| `events <resource> [id]` | Subscribe to real-time push events for a resource type, optionally scoped to an ID. Project and name filters are applied client-side; press Ctrl+C to disconnect. |
| `extension <action> <parent-resource> <parent-id> [id]` | Create, update, delete, retrieve, or list extensions attached to a project or resource. Actions that target one extension require its ID; create and update require a YAML file. IDs from YAML are retained unless reset, so reusing another parent's ID may affect that extension. Core replaces the complete extension list, so concurrent updates can overwrite one another. |
| `get <resource> [<id>]` | Retrieve a resource by ID or name. If no ID is supplied, use `--name` to retrieve the latest version; `--out` selects the output format. |
| `init [path]` | Initialize a Python virtual environment, write `.env`, install agent templates, and write a local `.dhcore.ini`, in that order. Use `--venv`, `--env-file`, `--agent`, and/or `--config` to select only specific steps. |
| `list <resource>` | List resources, optionally filtering by name, kind, or state. `--name` lists all versions matching that name. |
| `list-env` | List the environments registered in the CLI configuration. |
| `log <id>` | Read logs for a run. Supply its project; use `--container` to select a container and `--follow` to keep streaming logs. |
| `login` | Authenticate to an environment using OAuth2 PKCE, or provide a personal access token for a non-interactive login. |
| `metrics` | Read instance-wide metrics. Use `metrics project [name]` or `metrics run <id>` to scope metrics to a project or run. `--follow` refreshes every 15 seconds. |
| `port-forward [run-id]` | Start a local HTTP forward to a running service. Resolve the run by ID, function name, or run name; the project is required. |
| `proxy [run-id]` | Open a browser with authenticated access to a running service. Resolve the run by ID, function name, or run name; the project is required. Unlike port forwarding, browser traffic goes directly to the remote service. |
| `refresh` | Refresh the access token for the selected environment. |
| `register <endpoint>` | Register a Core endpoint as an environment. Use `--env` to name/select the environment and `--force` to replace an existing environment with a different endpoint. |
| `remove <environment>` | Remove a registered environment from the CLI configuration. |
| `run <task>` | Run a function/task in a project. Identify the function by name or ID, or provide its YAML definition. |
| `services` | List serving runs in a project. Filter by name, kind, or state; `--name` lists all versions. |
| `stop <id>` | Stop a run in the specified project. |
| `update <resource> <id>` | Update a resource from a YAML definition. The ID identifies the resource to update. |
| `upload <resource> [<id>]` | Upload a file or directory to S3. Provide an ID to update an existing resource; omit it to create a new artifact, for which a name is required. |
| `use <environment>` | Set the default registered environment used by subsequent commands. |

### Arguments and parameters

In command syntax, `<value>` is required and `[value]` is optional. `resource`
is the Core resource type; `id` identifies a resource or run. Commands acting on
resources inside a project generally require `--project` (`-p`), unless the
resource is a project or the project name is set in the `PROJECT_NAME`
environment variable.
The global `--env` (`-e`) flag selects a registered environment.

The following command-specific arguments and flags are available:

| Command | Arguments and flags |
| --- | --- |
| `config` | `-e, --env`: environment; `-o, --out`: output format (`short`, `json`, or `yaml`); `-p, --provider`: filter settings by provider (for example `dhcore`, `trino`, or `s3`). |
| `credentials` | `-e, --env`: environment; `-o, --out`: output format (`short`, `json`, or `yaml`); `-p, --provider`: filter credentials by provider. |
| `create <resource>` | `-e, --env`: environment; `-p, --project`: parent project for non-project resources; `-n, --name`: project name; `-f, --file`: YAML resource definition (required except when creating a project by name); `-r, --reset-id`: remove the YAML ID so Core assigns a new one. |
| `delete <resource> [<id>]` | `-e, --env`: environment; `-p, --project`: project for non-project resources; `-n, --name`: delete by name, including all versions; `-y, --confirm`: skip confirmation; `-c, --cascade`: also delete related resources when deleting a project. Supply either an ID or a name. |
| `download <resource> [<id>]` | `-e, --env`: environment; `-p, --project`: project for non-project resources; `-n, --name`: select the latest version by name; `-d, --destination`: output file or directory; `-o, --out`: output format (`short`, `json`, or `yaml`); `-v, --verbose`: show verbose progress/logging. |
| `events <resource> [id]` | `-e, --env`: environment; `-o, --out`: event format (`short`, `json`, or `yaml`); `-p, --project`: filter events by project; `-n, --name`: filter by resource name. |
| `extension <action> <parent-resource> <parent-id> [id]` | `action` is `create`, `update`, `delete`, `get`, or `list`. `parent-resource` can be `project`, `artifact`, `dataitem`, `function`, `model`, or `workflow`; `parent-id` is its identifier. Non-project parents require `-p, --project`. `get`, `update`, and `delete` require the final extension `id`. `-e, --env` selects an environment; `-f, --file` supplies the YAML definition for create/update; `-o, --out` formats get/list output (`yaml` by default; `short` or `json` also supported); `-r, --reset-id` removes the YAML ID on create so Core assigns one. Without `--reset-id`, a provided ID is retained. Updates retain the target ID and replace that extension's definition. |
| `get <resource> [<id>]` | `-e, --env`: environment; `-p, --project`: project for non-project resources; `-n, --name`: retrieve the latest version by name; `-o, --out`: output format (`short`, `json`, or `yaml`). |
| `init [path]` | `path`: virtual-environment location (default `.venv`), valid when the venv step is selected. `-e, --env`: registered environment to use; `--environment`: alternative environment selector, which overrides `--env` if both are provided. `--venv`, `--env-file`, `--agent`, and `--config` select the steps to run; without any of them all four run. `--pre` includes pre-release packages and requires the venv step. `--overwrite` updates an existing venv without prompting and overwrites/merges existing files for the selected steps. `--skip-gitignore` skips the final `.gitignore` update. The former `init --env` export flag has been renamed to `--env-file`. |
| `list <resource>` | `-e, --env`: environment; `-p, --project`: project for non-project resources; `-o, --out`: output format (`short` by default; `json` or `yaml` also supported); `-n, --name`: list all versions for a name; `-k, --kind` and `-s, --state`: filter by kind and state. |
| `log <id>` | `id`: run ID; `-e, --env`: environment; `-p, --project`: project; `-c, --container`: container ID; `-f, --follow`: keep streaming/refreshing logs. |
| `login` | `-e, --env`: environment; `--pat`: personal access token for non-interactive authentication. The token can also be supplied through `DHCORE_PERSONAL_ACCESS_TOKEN` or `DHCORE_PAT`. |
| `metrics` | `-e, --env`: environment; `-o, --out`: output format (`short`, `json`, or `yaml`); `-f, --follow`: refresh metrics every 15 seconds. |
| `metrics project [name]` | `name`: project name (optional if provided with `-p`); `-e, --env`: environment; `-p, --project`: project name; `-o, --out`: output format; `-f, --follow`: refresh every 15 seconds. |
| `metrics run <id>` | `id`: run ID; `-e, --env`: environment; `-p, --project`: project name (required); `-o, --out`: output format; `-f, --follow`: refresh every 15 seconds. |
| `port-forward [run-id]` | `run-id`: run ID, or identify the run with `-f, --function` or `-n, --name`; one of these selectors is required. `-e, --env`: environment; `-p, --project`: required project; `-l, --local-port`: local listening port (random if omitted). |
| `proxy [run-id]` | `run-id`: run ID, or identify the run with `-f, --function` or `-n, --name`; one of these selectors is required. `-e, --env`: environment; `-p, --project`: required project. |
| `refresh` | `-e, --env`: environment whose access token should be refreshed. |
| `register <endpoint>` | `endpoint`: Core instance URL; `-e, --env`: environment name; `-f, --force`: replace an existing environment if its endpoint differs. |
| `run <task>` | `task`: task/action to run; `-e, --env`: environment; `-p, --project`: project; `-n, --fn-name`: function name; `-i, --fn-id`: function ID; `-f, --file`: YAML function/resource definition. Use a function name or ID to select the function. |
| `services` | `-e, --env`: environment; `-p, --project`: project (required); `-o, --out`: output format (`short` by default; `json` or `yaml` also supported); `-n, --name`: list all versions of a service; `-k, --kind` and `-s, --state`: filter by kind and state. |
| `stop <id>` | `id`: run ID; `-e, --env`: environment; `-p, --project`: project. |
| `update <resource> <id>` | `resource` and `id`: resource type and target ID; `-e, --env`: environment; `-p, --project`: project for non-project resources; `-f, --file`: YAML definition to apply (required). |
| `upload <resource> [<id>]` | `resource` and optional `id`: resource type and existing resource ID; `-e, --env`: environment; `-p, --project`: project for non-project resources; `-f, --file`: required file or directory to upload; `-n, --name`: required when creating a new artifact without an ID; `-v, --verbose`: show verbose progress/logging. |
| `use <environment>` | `environment`: name of the registered environment to make the default. |
| `remove <environment>` | `environment`: name of the registered environment to remove. |
| `list-env` | No arguments or command-specific flags. |

For example, list a project's extensions, create an extension for a data item,
or retrieve one by its ID:

```sh
dhcli extension list project my-project
dhcli extension create -p my-project -f extension.yaml dataitem my-dataitem
dhcli extension get -p my-project dataitem my-dataitem extension-id
```

Root flags `-v, --verbose` enable verbose output and `--debug` enables HTTP
debug logging. Use `--help` on any command for the flags applicable to that
command.


## Version-pinned agent workspace

After registering and selecting a Core environment, run from your project directory:

```sh
dhcli init --agent
```

`--agent` fetches the active Core's `dhcore_version` from `/.well-known/configuration`,
downloads the latest agent templates from `scc-digitalhub/digitalhub-cli` (`main` branch),
and renders them into the current working directory (without creating a Python virtual
environment or installing Python SDK packages). If the remote archive is unavailable,
incomplete, or invalid, it falls back to the templates embedded in the CLI binary.
It also prints the template source and, when available, revision to stdout.
Core `0.16.2` selects the `0.16` knowledge branch.

```text
digitalhub-cli/templates/           User's working repository/
  AGENTS.md                          AGENTS.md
  CLAUDE.md                          CLAUDE.md
  skills/dh-*/SKILL.md                .agents/skills/dh-*/SKILL.md
```

The destination is the current working directory. `CLAUDE.md`
directs Claude to `AGENTS.md`, which links to the local, agent-neutral skills.
Every `<%DOCS%>` placeholder becomes `https://raw.githubusercontent.com/scc-digitalhub/docs/0.16`
in this example. Knowledge stays remote under that branch's `knowledge/` folder and is
fetched by the agent when needed. Relative knowledge links resolve against the fetched
page URL, keeping the same release. Agents must report unavailable knowledge rather than
fall back to `main`, `latest`, or a different release.

Identical files are left untouched on repeated initialization. By default, differing existing
files are never overwritten; pass `dhcli init --agent --overwrite` to update existing `AGENTS.md`,
`CLAUDE.md`, and skill files in place when remote templates are fetched (back up customizations
before running with `--overwrite`). When falling back to embedded templates (for example, when
GitHub is unreachable), only missing template files are created and existing files are preserved
even with `--overwrite` so an offline run never downgrades newer remote templates to older
embedded versions. Symlink destinations are rejected. Failed writes attempt to roll back files
and directories created or modified by that installation; any rollback errors are reported.
Without any step-selection flags, `dhcli init [path]` initializes the venv, `.env`, agent
templates, and local `.dhcore.ini` in that order. Use `dhcli init --venv [path]` for only the Python environment, or combine
selection flags such as `dhcli init --env-file --agent` to run only those steps. The command stops
on the first error; completed steps are not rolled back.
An existing venv is reused after confirmation; `--overwrite` skips the prompt and upgrades the
selected DigitalHub packages in place without recreating the environment. Dependencies may also change.
Use `dhcli init --config` to write only the currently selected environment to `.dhcore.ini` in
the current directory. The file may contain credentials, is owner-readable only, and should not be
committed. An existing `.dhcore.ini` is replaced entirely only with `--overwrite`, removing any
other environments in that file. `DH_CONFIG`, if set, takes precedence over the local file.
As the final step, `init` creates or updates `.gitignore` with the output paths of the selected
steps. Agent initialization adds `AGENTS.md`, `CLAUDE.md`, and `.agents/skills/`; unrelated paths
are not added. A venv
outside the current directory cannot be ignored by this `.gitignore` and is reported for manual
handling. If a step fails, `.gitignore` is not updated. Use `--skip-gitignore` to leave it
untouched. Ignore rules do not remove files already tracked by Git.

Maintain the router and skills in [`templates/`](templates/); `go:embed` includes them in every
CLI build and release without external asset files. Publish release-specific knowledge under
`knowledge/` on the corresponding [`scc-digitalhub/docs`](https://github.com/scc-digitalhub/docs)
branch. Do not copy templates back into the docs repository. Run `go vet ./...` and
`go build ./...` before release.

## Development

- `core/commands` contains the definition of available commands and what flags they accept
- `core/flags` contains the generic flag handling system
- `core/facade` contains the business logic handlers for each command
- `core/facade/adapter` bridges CLI commands to the SDK services

### Build from source

If you wish to build the executable from source, run the following:

```sh
go build
```

To build with a specific version:

```sh
go build -ldflags "-X dhcli/core.version=0.15.0"
```

It will generate an executable named `dhcli` for your operating system and architecture. To change the target OS or architecture, you need to set the `GOOS` and `GOARCH` variables and build it. Some examples:

```sh
GOOS=linux GOARCH=amd64 go build -o dhcli-linux-amd64
```

```sh
GOOS=darwin GOARCH=arm64 go build -o dhcli-darwin-arm64
```

```sh
GOOS=windows GOARCH=amd64 go build -o dhcli-win-amd64.exe
```

For a complete list of available values, run `go tool dist list`, which will return the list of valid OS/ARCH combinations.

### Publish a new release

A GitHub Actions workflow is set to be triggered upon pushing a new tag and will automatically create a new release through GoReleaser. The version is injected automatically from the tag.

Tags must use the `v` prefix (required by Go modules):

```sh
git tag -a vX.Y.Z -m "Some release" && git push origin vX.Y.Z
```

A versioned Homebrew formula (`dhcli@X.Y.rb`) is also created automatically.

## Security Policy

The current release is the supported version. Security fixes are released together with all other fixes in each new release.

If you discover a security vulnerability in this project, please do not open a public issue.

Instead, report it privately by emailing us at digitalhub@fbk.eu. Include as much detail as possible to help us understand and address the issue quickly and responsibly.

## Contributing

To report a bug or request a feature, please first check the existing issues to avoid duplicates. If none exist, open a new issue with a clear title and a detailed description, including any steps to reproduce if it's a bug.

To contribute code, start by forking the repository. Clone your fork locally and create a new branch for your changes. Make sure your commits follow the [Conventional Commits v1.0](https://www.conventionalcommits.org/en/v1.0.0/) specification to keep history readable and consistent.

Once your changes are ready, push your branch to your fork and open a pull request against the main branch. Be sure to include a summary of what you changed and why. If your pull request addresses an issue, mention it in the description (e.g., “Closes #123”).

Please note that new contributors may be asked to sign a Contributor License Agreement (CLA) before their pull requests can be merged. This helps us ensure compliance with open source licensing standards.

We appreciate contributions and help in improving the project!

## Authors

This project is developed and maintained by **DSLab – Fondazione Bruno Kessler**, with contributions from the open source community. A complete list of contributors is available in the project’s commit history and pull requests.

For questions or inquiries, please contact: [digitalhub@fbk.eu](mailto:digitalhub@fbk.eu)

## Copyright and license

Copyright © 2025 DSLab – Fondazione Bruno Kessler and individual contributors.

This project is licensed under the Apache License, Version 2.0.
You may not use this file except in compliance with the License. Ownership of contributions remains with the original authors and is governed by the terms of the Apache 2.0 License, including the requirement to grant a license to the project.
