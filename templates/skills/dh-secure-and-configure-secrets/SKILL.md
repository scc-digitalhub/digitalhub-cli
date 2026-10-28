---
name: dh-secure-and-configure-secrets
description: Manage Kubernetes-backed project secrets in DigitalHub, create secrets using the Python SDK fallback, and securely inject credentials into Function and Workflow runs via spec.secrets.
---

# Skill: Managing Secrets & Secure Credential Injection

Use this procedural workflow whenever workloads require sensitive credentials (HuggingFace API tokens for gated models like Llama-3, database passwords, object storage keys, or external API secrets).

## Reference Knowledge (Pinned to Platform Version)
- **Secret Entity Specification**: [secret.md]({{DOCS}}/knowledge/entities/secret.md)
- **Kubernetes Resources & Secrets**: [kubernetes-resources.md]({{DOCS}}/knowledge/management/kubernetes-resources.md)
- **Platform Governance & Secret Rules**: [governance-and-security.md]({{DOCS}}/knowledge/management/governance-and-security.md)
- **Python SDK Fallback Guide**: [sdk-fallback.md]({{DOCS}}/knowledge/sdk/sdk-fallback.md)

---

## 1. Pre-Flight Security Gate
- [ ] **Zero-Plaintext Rule**: **NEVER** write passwords, tokens, or private keys inside YAML manifests (`spec.envs`), scripts (`handler.py`), or git commits.
- [ ] Secret names are injected as environment variable names: use `UPPER_SNAKE_CASE` by convention (e.g. `HF_TOKEN`, `DB_PASSWORD`). Core only enforces the `secret://([\w-]+)` path pattern.
- [ ] Target project exists (`dhcli get projects <project> -o json`).

---

## 2. Step-by-Step Procedure

### Step 1: Create or Update a Project Secret
Because `secret` is not in `dhcli`'s positional CRUD, create it via the verified Python SDK fallback:

```bash
python3 -c '
import digitalhub as dh
p = dh.get_project("<project>")
p.new_secret(name="<SECRET_NAME>", secret_value="<SECRET_VALUE>")
print("Secret <SECRET_NAME> created successfully in project <project>")
'
```

### Step 2: Mount the Secret into a Run Manifest (`spec.secrets`)
In your `run.yaml` (or Function default spec), declare the secret name in the `spec.secrets` array:

```yaml
spec:
  inputs:
    # ... inputs ...
  secrets:
    - <SECRET_NAME>
  resources:
    cpu: "1"
    mem: "2Gi"
```

### Step 3: Consume the Secret Inside In-Container Code
DigitalHub injects each declared secret as a native environment variable matching its name:

```python
import os

def my_handler(project, run):
    secret_token = os.environ.get("<SECRET_NAME>")
    if not secret_token:
        raise ValueError("Missing mandatory secret: <SECRET_NAME>")
    # Use secret_token to authenticate with external APIs or databases
```

### Step 4: Verification
Launch the run via `dhcli run` and verify the container boots without credential authentication errors:
```bash
dhcli run python+job -p <project> -n <function-name> -f run.yaml
```
Check logs if needed:
```bash
dhcli log <run-id> -p <project>
```
