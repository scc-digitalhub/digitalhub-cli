---
name: dh-orchestrate-workflow
description: Author multi-step DAG pipelines in Python using Hera (Argo) or KFP DSL, compile and execute workflows via the SDK fallback, and schedule automated recurring triggers.
---

# Skill: Orchestrating Workflow DAGs & Automated Triggers

Use this procedural workflow whenever chaining multiple DigitalHub Functions into automated pipelines using Hera (Argo Workflows) or Kubeflow Pipelines (KFP), and scheduling runs with cron (`scheduler`) or event-driven (`lifecycle`) Triggers.

## Reference Knowledge (Pinned to Platform Version)
- **Hera Workflow Runtime**: [hera.md](<%DOCS%>/knowledge/runtimes/hera.md)
- **KFP Workflow Runtime**: [kfp.md](<%DOCS%>/knowledge/runtimes/kfp.md)
- **Workflow Entity Specification**: [workflow.md](<%DOCS%>/knowledge/entities/workflow.md)
- **Trigger Entity Specification**: [trigger.md](<%DOCS%>/knowledge/entities/trigger.md)
- **Python SDK Fallback Guide**: [sdk-fallback.md](<%DOCS%>/knowledge/sdk/sdk-fallback.md)

---

## 1. Pre-Flight Checklist
- [ ] Target `<project>` exists (`dhcli get projects <project> -o json`).
- [ ] All upstream Functions called by the DAG are already registered and in `READY` state (`dhcli list functions -p <project>`).
- [ ] Local virtual environment initialized via `echo y | dhcli init ./.venv` and activated.

---

## 2. Step-by-Step Procedure

### Step 1: Author the Python Hera Pipeline Script (`pipeline.py`)
```python
from digitalhub_runtime_hera.dsl import step
from hera.workflows import Workflow, DAG, Parameter

def pipeline():
    with Workflow(entrypoint="dag", arguments=[Parameter(name="dataset_key")]) as w:
        with DAG(name="dag"):
            s1 = step(
                name="prep",
                function="prep-fn",
                template={"action": "job", "inputs": {"raw": "{{workflow.parameters.dataset_key}}"}},
                outputs=["clean_key"]
            )
            s2 = step(
                name="train",
                function="train-fn",
                template={"action": "job", "inputs": {"train_data": "{{inputs.parameters.clean_key}}"}},
                arguments={"clean_key": s1.get_parameter("clean_key")}
            )
            s1 >> s2
    return w
```

Encode the pipeline:
```bash
base64 -i pipeline.py | tr -d '\n'
```

### Step 2: Register the Workflow Entity via `dhcli create`
```yaml
kind: hera
name: training-pipeline
project: <project>
metadata:
  description: "End-to-end data prep and model training DAG"
spec:
  source:
    lang: python
    handler: pipeline
    base64: "<BASE64_ENCODED_PIPELINE>"
```
```bash
dhcli create workflows -p <project> -f workflow.yaml
```

### Step 3: Build & Execute the Pipeline (SDK Fallback)
Because `dhcli run` does not target `/workflows`, run `build` and `pipeline` via the one-line SDK fallback:

```bash
python3 -c '
import digitalhub as dh
wf = dh.get_project("<project>").get_workflow("training-pipeline")
wf.run("build", wait=True)
run = wf.run("pipeline", parameters={"dataset_key": "store://<project>/dataitem/table/dataset:latest"})
print("Pipeline Run ID:", run.id)
'
```

### Step 4: Attach a Cron or Lifecycle Trigger (Optional)
To schedule recurring runs (using 6-field Quartz cron):
```bash
python3 -c '
import digitalhub as dh
wf = dh.get_project("<project>").get_workflow("training-pipeline")
wf.trigger(
    action="pipeline",
    name="nightly-run",
    kind="scheduler",
    schedule="0 0 2 * * ?",
    parameters={"dataset_key": "store://<project>/dataitem/table/dataset:latest"},
)
print("Trigger attached successfully.")
'
```

### Step 5: Observability
Monitor the pipeline run state and steps:
```bash
dhcli list runs -p <project> -o json
dhcli get runs <run-id> -p <project> -o json | jq 'del(.status.k8s)'
```
