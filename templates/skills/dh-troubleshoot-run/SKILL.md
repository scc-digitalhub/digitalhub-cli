---
name: dh-troubleshoot-run
description: Diagnose and fix failed DigitalHub Runs (ERROR state), YAML schema validation errors, Kaniko container build failures, Python handler exceptions, and Kubernetes pod crashes using dhcli.
---

# Skill: Troubleshooting Failed Runs & CLI Errors

Use this procedural workflow whenever a `dhcli` command returns an error or a DigitalHub Run (`python+job:run`, `python+build:run`, `hera+pipeline:run`, `*+serve:run`) enters `status.state: ERROR` or gets stuck in `READY`/`RUNNING`.

## Reference Knowledge (Pinned to Platform Version)
- **Diagnostic & Troubleshooting Runbook**: [troubleshooting.md](<%DOCS%>/knowledge/management/troubleshooting.md)
- **Run Entity & State Machine**: [run.md](<%DOCS%>/knowledge/entities/run.md)
- **Kubernetes Compute, Memory & Volumes**: [kubernetes-resources.md](<%DOCS%>/knowledge/management/kubernetes-resources.md)

---

## 1. Systematic 3-Step Triage Protocol

Whenever a Run fails or stalls:

### Step 1: Inspect the Run Status (Exclude `.status.k8s` to Avoid Truncation)
```bash
dhcli get runs <run-id> -p <project> -o json | jq 'del(.status.k8s)'
```
Identify the failure phase:
- Failed immediately before pod creation → **Schema / Task / Entity Reference Error**.
- Stuck in `READY` or failed during startup (`ImagePullBackOff`, `OOMKilled`) → **K8s Resource / Image Error**.
- Container ran and exited non-zero → **Application / Python Handler / Build Error**.

### Step 2: Retrieve Pod Logs (Live Container or Persisted Logs)
1. Try fetching live container logs (never pass `-f` in an automated tool call):
   ```bash
   dhcli log <run-id> -p <project>
   ```
2. **If `dhcli log` fails with `Failed: container not found`** (common for terminated or completed pods):
   Fetch the persisted logs directly from DigitalHub Core's log store:
   ```bash
   dhcli list logs -p <project> -o json | jq -r '.[] | select(.run == "<run-id>") | .content | @base64d'
   ```

---

## 2. Match Symptom to Root Cause & Apply Fix

### Case A: Python Handler Exception (`ModuleNotFoundError`, `TypeError`, `KeyError`)
1. **`ModuleNotFoundError: No module named '<pkg>'`**:
   - Inspect the function: `dhcli get functions -p <project> -n <fn-name> -o yaml`
   - Add `<pkg>` to `spec.requirements`, create a new function version, run `python+build` until `COMPLETED`, and re-run.
2. **Handler Code Bug Fix (FAST-RETRY via Image Reuse)**:
   - When only fixing Python handler code (and `spec.requirements` did not change):
   - Retrieve the previously built image:
     ```bash
     dhcli get functions -p <project> -n <fn-name> -o json | jq -r '.spec.image'
     ```
   - Base64-encode the corrected handler (`base64 -i handler.py | tr -d '\n'`), copy `spec.image: "<built-image-tag>"` into your new `function.yaml`, register the new function version, and re-run `python+job` **immediately without waiting for another build**!
3. **`TypeError: unexpected keyword argument`**:
   - Align the keys in `run.yaml`'s `spec.inputs` and `spec.parameters` with the argument names of `def handler(...):`.

### Case B: Kaniko Image Build Failure (`python+build:run` → `ERROR`)
1. Inspect the build logs via the persisted log command above.
2. Fix invalid package versions in `spec.requirements` or pass OS-level apt commands in `run.yaml` under `spec.instructions: ["RUN apt-get update && apt-get install -y gcc libpq-dev"]`.

### Case C: Kubernetes `OOMKilled` (Exit Code 137) or Resource Starvation
1. If `OOMKilled`: Increase `spec.resources.mem` (e.g., `"8Gi"`) in `run.yaml`, and mount an in-memory `/dev/shm` `empty_dir` volume when running PyTorch, HuggingFace, or vLLM.
2. If stuck in `PENDING` waiting for GPU/Memory: Lower `spec.resources` or check the cluster hardware `spec.profile`.

---

## 3. Re-Run & Verify Recovery
Trigger a new run with `dhcli run` and confirm it reaches `COMPLETED` (for jobs/builds) or `RUNNING` (for services):
```bash
dhcli run <kind>+<action> -p <project> -n <fn-name> -f fixed-run.yaml
dhcli get runs <new-run-id> -p <project> -o json | jq 'del(.status.k8s)'
```
