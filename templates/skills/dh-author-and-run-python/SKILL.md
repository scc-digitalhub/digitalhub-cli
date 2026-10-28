---
name: dh-author-and-run-python
description: Author a DigitalHub Python function (batch job or Nuclio HTTP serve), Base64-encode the handler script, build custom dependencies (python+build), execute via dhcli run (python+job / python+serve), and inspect logs.
---

# Skill: Authoring, Building & Running Python Functions

Use this procedural workflow whenever creating or updating a Python Function (`kind: python`), building custom `requirements` with `python+build`, running batch jobs (`python+job`) or Nuclio HTTP endpoints (`python+serve`), and verifying execution logs.

## Reference Knowledge (Pinned to Platform Version)
- **Python Runtime Specification**: [python.md]({{DOCS}}/knowledge/runtimes/python.md)
- **Function & Run Entities**: [function.md]({{DOCS}}/knowledge/entities/function.md), [run.md]({{DOCS}}/knowledge/entities/run.md)
- **Writing Python Handler Code**: [handler-authoring.md]({{DOCS}}/knowledge/sdk/handler-authoring.md)
- **Reading & Logging Data/Models/Metrics**: [data-models-and-metrics.md]({{DOCS}}/knowledge/sdk/data-models-and-metrics.md)

---

## 1. Pre-Flight Checklist
- [ ] Target `<project>` exists (`dhcli get projects <project> -o json`).
- [ ] Function entity name uses RFC-1123 (`^[a-z0-9-]+$`, max 63 chars).
- [ ] Python handler name uses `snake_case`.

---

## 2. Step-by-Step Procedure

### Step 1: Author & Base64-Encode the Python Handler Script
1. Write a clean Python script (`handler.py`) containing your entrypoint function:
   - **Injected arguments**: Declare `project` and/or `run` in your function signature to log outputs (`project.log_dataitem(...)`, `project.log_artifact(...)`, `project.log_sklearn(...)`, `run.log_metric(...)`).
   - **Inputs vs. Parameters**: Arguments passed via `spec.inputs` arrive as SDK `Dataitem`/`Artifact`/`Model` objects (call `.as_df()` or `.download()`). Arguments passed via `spec.parameters` arrive as plain Python primitives (`str`, `int`, `float`, `bool`, `dict`, `list`).
   - **For `python+serve`**: Define `def init_context(context, ...):` (optional startup hook) and `def serve(context, event):` (request handler).
2. **Base64-encode the local `.py` file** into a single line:
   ```bash
   base64 -i handler.py | tr -d '\n'
   ```

### Step 2: Create the Function Entity via `dhcli create`
Create `function.yaml` with `lang`, `handler`, and `base64` nested under `spec.source`:
```yaml
kind: python
name: <function-name>
project: <project>
metadata:
  description: "<brief description>"
spec:
  python_version: PYTHON3_10
  requirements:
    - "<pip-package-1>"
  source:
    lang: python
    handler: <entrypoint_function_name>
    base64: "<single-line-base64-encoded-code>"
```
Register the function in DigitalHub Core:
```bash
dhcli create functions -p <project> -f function.yaml
dhcli get functions -p <project> -n <function-name> -o json
```

> ⚡ **FAST-RETRY (Image Reuse Optimization)**:
> If you are updating/fixing the Python handler code and `spec.requirements` did not change (and a previous `python+build` already finished), retrieve the existing image:
> ```bash
> dhcli get functions -p <project> -n <function-name> -o json | jq -r '.spec.image'
> ```
> Add `spec.image: "<image-tag>"` directly to your new `function.yaml`. This lets you launch `python+job` immediately **without rebuilding**!

### Step 3: Build Dependencies (If Needed) & Run via `dhcli run`
1. **When `spec.requirements` has third-party packages** (and not reusing an image), run `python+build` first and wait until `COMPLETED`:
   ```bash
   dhcli run python+build -p <project> -n <function-name>
   ```
2. **Prepare `run.yaml`**:
   > **Note**: Do **not** write `spec.function` or `spec.task` in `run.yaml`—`dhcli run` resolves those automatically.
   ```yaml
   spec:
     inputs:
       # handler_arg: "store://<project>/<entity-type>/<kind>/<name>:<uuid>"
     parameters:
       # handler_arg: value
     resources:
       cpu: "1"
       mem: "2Gi"
   ```
3. **Execute the run**:
   - Batch job:
     ```bash
     dhcli run python+job -p <project> -n <function-name> -f run.yaml
     ```
   - Serverless HTTP service:
     ```bash
     dhcli run python+serve -p <project> -n <function-name> -f run.yaml
     ```

### Step 4: Monitor State & Inspect Logs
1. Retrieve the newly created `<run-id>` and check `status.state`:
   ```bash
   dhcli list runs -p <project> -o json
   dhcli get runs <run-id> -p <project> -o json | jq 'del(.status.k8s)'
   ```
2. Fetch container logs (live pod) or persisted logs (terminated pod):
   ```bash
   dhcli log <run-id> -p <project>
   # If returns "Failed: container not found":
   dhcli list logs -p <project> -o json | jq -r '.[] | select(.run == "<run-id>") | .content | @base64d'
   ```
3. If `status.state` is `ERROR`, inspect the traceback, fix `handler.py`, apply the Image Reuse optimization, create a new function version, and re-run.
