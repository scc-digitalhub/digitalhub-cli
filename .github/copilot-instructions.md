# DigitalHub CLI - AI Agent Instructions

## Project Overview
DigitalHub CLI (`dhcli`) is a Go-based command-line tool for managing DigitalHub platform resources. It provides CRUD operations for entities like projects, artifacts, functions, models, runs, and workflows via REST API calls.

## Architecture
- **Entry Point**: `main.go` initializes and executes the root Cobra command.
- **Command Layer**: `core/commands/` defines CLI commands using Cobra, each registering via `core.RegisterCommand()`.
- **Facade Layer**: `core/facade/` contains business logic handlers for commands.
- **Adapter Layer**: `core/facade/adapter/` integrates with `digitalhub-cli-sdk` for API operations, handling authentication, config translation, and CRUD calls.
- **Flags**: `core/flags/` provides a custom generic flag system wrapping Cobra flags.
- **Config**: Uses Viper + ini files for configuration, loaded via SDK utils.

## Key Patterns
- **Command Registration**: Commands are defined as functions returning `*cobra.Command`, registered in `init()` blocks.
- **Error Handling**: Use `log.Fatalf()` for fatal errors, `os.Exit(1)` for validation failures in adapters.
- **Config Access**: Always use `viper.GetString()` for config values; adapter translates Viper config to SDK config structs.
- **Endpoint Translation**: Use `utils.TranslateEndpoint(resource)` to map resource types to API endpoints.
- **Version Checks**: Call `utils.CheckApiLevel()` and `utils.CheckUpdateEnvironment()` before API operations.
- **Copyright Headers**: Include SPDX headers: `// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler\n//\n// SPDX-License-Identifier: Apache-2.0`

## Development Workflow
- **Build**: `go build` produces `dhcli` executable. Cross-compile with `GOOS=linux GOARCH=amd64 go build -o dhcli-linux-amd64`.
- **Dependencies**: Managed via `go.mod`; key external: `digitalhub-cli-sdk` for API, `charmbracelet/fang` for CLI execution.
- **Release**: Push semantic version tags (e.g., `v1.2.3`) to trigger GoReleaser workflow.
- **Testing**: No automated tests present; validate manually via CLI commands.
- **Local Setup**: Requires Python 3.9-3.12 for `init` command; installs DigitalHub packages via pip.

## Code Examples
- **Adding a Command**: Create `core/commands/newcmd.go` with command definition and `facade.NewCmdHandler()` implementation.
- **API Call**: In adapter, create SDK config from Viper, use `crudsvc.Create()` with context.
- **Flag Usage**: Define flags with `flags.NewStringFlag()`, add to command with `flags.AddFlag(cmd, &flag)`.

## Common Pitfalls
- Avoid direct API calls in facade; delegate to adapter layer.
- Ensure project context for non-project resources.
- Handle config loading failures in `PersistentPreRunE`.</content>
<parameter name="filePath">/home/ltrubbiani/GolandProjects/digitalhub-cli/.github/copilot-instructions.md