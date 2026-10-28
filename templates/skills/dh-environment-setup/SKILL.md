---
name: dh-environment-setup
description: Configure and verify the DigitalHub CLI (dhcli) environment in ~/.dhcore.ini, register Core endpoints, authenticate via OAuth2 PKCE or PAT, bootstrap a Python SDK venv with dhcli init, and create or inspect DigitalHub projects.
---

# Skill: DigitalHub Environment Setup & Project Initialization

Use this procedural workflow whenever starting work in a workspace, registering or switching DigitalHub environments, refreshing expired OAuth2 tokens, bootstrapping a local Python SDK virtual environment, or creating a DigitalHub project.

## Reference Knowledge (Pinned to Platform Version)
- **CLI Auth & Environments**: [auth-and-env.md]({{DOCS}}/knowledge/cli/auth-and-env.md)
- **Project Entity Specification**: [project.md]({{DOCS}}/knowledge/entities/project.md)
- **CLI CRUD & YAML Anatomy**: [crud-and-yaml-anatomy.md]({{DOCS}}/knowledge/cli/crud-and-yaml-anatomy.md)
- **Governance & Naming Rules**: [governance-and-security.md]({{DOCS}}/knowledge/management/governance-and-security.md)

---

## 1. Pre-Flight Checklist
- [ ] Project names **must** obey RFC-1123 (`^[a-z0-9-]+$`, lowercase alphanumeric and hyphens, max 63 chars, e.g., `fraud-detection`).
- [ ] `dhcli` is installed and on `PATH` (`dhcli --version`).

---

## 2. Step-by-Step Procedure

### Step 1: Verify `dhcli` Installation
Check that the `dhcli` binary is installed:
```bash
dhcli --version
```

### Step 2: Inspect & Select the Active Environment (`~/.dhcore.ini`)
1. List all registered environments in `~/.dhcore.ini`:
   ```bash
   dhcli list-env
   ```
2. Inspect the non-secret configuration (`dhcore_endpoint`, `dhcore_version`, `s3_bucket`) of the active environment:
   ```bash
   dhcli config -o json
   ```
3. To switch to an existing registered environment:
   ```bash
   dhcli use <env-name>
   ```
4. To register a new DigitalHub Core endpoint (`/.well-known/configuration` is fetched automatically):
   ```bash
   dhcli register -e <env-name> https://<dhcore-endpoint>
   dhcli use <env-name>
   ```

### Step 3: Verify Authentication & Refresh Tokens
Test connectivity and authentication (use compact table output to avoid large JSON dumps):
```bash
dhcli list projects
```
- If the access token is expired (`401 Unauthorized`), try refreshing it first:
  ```bash
  dhcli refresh
  ```
- If a fresh login is required:
  - **Headless / Agent / CI environments (Personal Access Token)**:
    ```bash
    dhcli login --pat "<personal-access-token>"
    # Or export DHCORE_PERSONAL_ACCESS_TOKEN="<pat>" and run: dhcli login
    ```
  - **Interactive desktop environments (OAuth2 PKCE browser flow)**:
    ```bash
    dhcli login
    ```

### Step 4: Create or Verify the Target Project
1. Check if the target project already exists:
   ```bash
   dhcli get projects <project> -o json
   ```
2. If it does not exist, create it directly by name:
   ```bash
   dhcli create projects -n <project>
   ```
   *(Or from a manifest via `dhcli create projects -f project.yaml`.)*

### Step 5: Inspect Project Contents & Optional Python Venv Bootstrap
1. List existing entities in the project:
   ```bash
   dhcli list functions -p <project>
   dhcli list dataitems -p <project>
   dhcli list models -p <project>
   dhcli list workflows -p <project>
   ```
2. *(Optional)* To install version-pinned agent instructions and skills (`AGENTS.md`, `CLAUDE.md`, and `.agents/skills/`) in the current working directory:
   ```bash
   dhcli init --agent
   ```
   This fetches the active Core's version and renders the CLI's embedded templates with knowledge references to the matching `major.minor` docs branch (without creating a Python virtual environment). Identical agent files are left unchanged; conflicting files are reported as errors rather than overwritten.
3. *(Optional)* If you need a local Python virtual environment with `digitalhub[full]` matching the connected Core server's minor version (for example, to run SDK gap fallbacks or author interactive notebooks):
   ```bash
   echo y | dhcli init ./.venv
   ```
