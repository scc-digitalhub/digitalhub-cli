# CLI

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/scc-digitalhub/digitalhub-cli/release.yaml?event=release) [![license](https://img.shields.io/badge/license-Apache%202.0-blue)](https://github.com/scc-digitalhub/digitalhub-cli/LICENSE) ![GitHub Release](https://img.shields.io/github/v/release/scc-digitalhub/digitalhub-cli)
![Status](https://img.shields.io/badge/status-stable-gold)


The Command-Line Interface (CLI) is a simple tool that offers a number of functionalities to set the platform up.

The executable can be downloaded in the *Releases* page.

For a full overview of available commands, please check [the documentation](https://scc-digitalhub.github.io/docs/0.11/cli/commands/).

## Quick start

Download and unpack the corresponding archive from the [GitHub release page](https://github.com/scc-digitalhub/digitalhub-cli/releases). Run the ``dhcli`` executable. 

On Mac, you can use the *Homebrew Tap* distribution:

``` sh
brew tap scc-digitalhub/digitalhub-cli https://github.com/scc-digitalhub/digitalhub-cli
brew install dhcli
```

## Development

- `core/commands` contains the definition of available commands and what flags they accept
- `core/flags` lists common flags
- `core/service` contains the actual logic to handle each command
- `utils` contains common methods and values

### Build from source

If you wish to build the executable from source, run the following:

``` sh
go build
```

It will generate an executable named `dhcli` for your operating system and architecture. To change the target OS or architecture, you need to set the `GOOS` and `GOARCH` variables and build it. Some examples:
``` sh
GOOS=linux GOARCH=amd64 go build -o dhcli-linux-amd64
```
``` sh
GOOS=darwin GOARCH=arm64 go build -o dhcli-darwin-arm64
```
``` sh
GOOS=windows GOARCH=amd64 go build -o dhcli-win-amd64.exe
```

For a complete list of available values, run `go tool dist list`, which will return the list of valid OS/ARCH combinations.

### Publish a new release

A GitHub Actions workflow is set to be triggered upon pushing a new tag and will automatically create a new release through GoReleaser.

Simply push a new tag to have the corresponding release generated:

``` sh
git tag -a X.Y.Z -m "Some release" && git push origin X.Y.Z
```

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
