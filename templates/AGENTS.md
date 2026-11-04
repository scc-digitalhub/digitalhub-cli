# AGENTS.md — DigitalHub Platform Agent Router & Operating Manual

Welcome! This file is the **Single Source of Truth Router** for AI Coding Assistants (Claude Code, GitHub Copilot, Cursor, Windsurf, Codex, Antigravity) and the **DigitalHub Platform Agent (`digitalhub-agent`)**.

- **Versioned Docs Root**: `<%DOCS%>`. `dhcli init --agent` fetches the connected Core's `dhcore_version` from `/.well-known/configuration` and selects its `major.minor` docs branch (e.g., `0.16.2` selects `0.16`). The router and skills are embedded in the CLI; platform knowledge is fetched on demand from this root's `knowledge/` folder. Never fall back to `main`, `latest`, or a different release if knowledge is unavailable.
- **Following Links**: Resolve relative links in a fetched knowledge page against that page's full URL, not the user's filesystem. For example, `../runtimes/sklearnserve.md` in `<%DOCS%>/knowledge/entities/model.md` resolves to `<%DOCS%>/knowledge/runtimes/sklearnserve.md`. Catalog `path` values are relative to `<%DOCS%>/knowledge/`. Knowledge pages are served as-is and do not contain substitution placeholders.
- **Machine-Readable Catalog**: [`<%DOCS%>/knowledge/llms.txt`](<%DOCS%>/knowledge/llms.txt) and [`<%DOCS%>/knowledge/index.md`](<%DOCS%>/knowledge/index.md).

---

## 1. Golden Architectural Rules (Read Before Taking Any Action)

1. **CLI-First Execution (`dhcli` + Core JSON Schemas)**:
   - All host-side platform management, entity CRUD (`projects`, `functions`, `workflows`, `runs`, `dataitems`, `artifacts`, `models`, `logs`), S3 file transfers (`upload`, `download`), task execution (`run`), log inspection (`log`), and service exposure (`services`, `port-forward`) **MUST be performed using `dhcli` and declarative YAML manifests**.
   - Use the Python SDK (`import digitalhub as dh`) on the host **only** for the three CLI gap fallbacks documented in [`sdk-fallback.md`](<%DOCS%>/knowledge/sdk/sdk-fallback.md) (running a `workflow` pipeline, creating a `secret`, or attaching a `trigger`), or when the user explicitly asks for an interactive Jupyter Notebook (`.ipynb`).
2. **Strict Separation Between CLI (`dhcli`) and In-Container Python SDK (`digitalhub`)**:
   - **Host Side (`dhcli`)**: Registers environments (`register`, `login`, `use`), creates projects and entities (`create`), uploads/downloads S3 files (`upload`, `download`), triggers function runs (`dhcli run <kind>+<action> -p <project> -n <fn-name> [-f run.yaml]`), and reads logs (`dhcli log`).
   - **In-Container Side (`digitalhub` Python SDK)**: Used **inside** `.py` handler scripts (`def handler(project, run, ...):`) and pipeline scripts (`def pipeline():`) to read inputs (`di.as_df()`, `art.download()`) and log outputs (`project.log_dataitem()`, `project.log_sklearn()`, `project.log_mlflow()`, `project.log_artifact()`, `run.log_metric()`).
3. **Verified `dhcli` Command Rules**:
   - **No `dhcli execute` command exists**: Always trigger function runs via `dhcli run <kind>+<action> -p <project> -n <fn-name> [-f run.yaml]`.
   - **How `dhcli run` processes `-f run.yaml`**: `dhcli run` extracts **only the `spec:` block** from `run.yaml`, automatically resolves/creates the `Task`, and injects `spec.task`, `spec.function`, and `spec.local_execution: false`. Do **not** hand-write `spec.function` or `spec.task` inside `run.yaml` when using `dhcli run`.
   - **`dhcli upload` requires singular resource names**: When uploading a new entity (`dhcli upload <resource> -p <project> -n <name> -f <path>`), always pass the **singular** resource name (`artifact`, `dataitem`, or `model`) because `dhcli upload` sets the entity's `kind` to the `<resource>` argument.
   - **Never pass `-f` (`--follow`) to `dhcli log` or `dhcli metrics`** in an automated agent tool call—it blocks indefinitely.
4. **Critical YAML Schema Rules (`digitalhub-core/schemas/specs/`)**:
   - **Nested `spec.source`**: In code-backed Functions (`python`, `dbt`, `openinference`, `guardrail`, `flower-*`) and Workflows (`hera`, `kfp`), source fields (`lang`, `handler`, `base64`, `source`, `init_function`) **MUST** be nested under `spec.source`.
   - **Base64 Encoding for Local Scripts**: Because `dhcli create` sends YAML directly to DigitalHub Core without reading local `.py` or `.sql` files from disk, always Base64-encode local files (`base64 -i handler.py | tr -d '\n'`) into `spec.source.base64`.
   - **Input Entity Format (`spec.inputs`)**: Inside `run.yaml`, reference input Dataitems, Artifacts, and Models using full `store://<project>/<entity-type>/<kind>/<name>:<uuid>` URIs.
   - **RFC-1123 Naming**: All entity `name` and `project` fields must match `^[a-z0-9-]+$` (lowercase alphanumeric and hyphens, max 63 chars). Python handler names (`spec.source.handler`) must use `snake_case`.

---

## 2. Procedural Agent Skills (`.agents/skills/dh-*/SKILL.md`)

When executing a multi-step platform task, load and follow the corresponding procedural **Skill** runbook first. The CLI embeds this router and all skills from its `templates/` directory. `dhcli init --agent` installs skills into `.agents/skills/` in the current working directory, alongside `AGENTS.md` and a `CLAUDE.md` pointer. The links below refer to the installed local skills; only their knowledge references are remote and version-pinned:

| Skill Name | Path | When to Trigger |
| :--- | :--- | :--- |
| **`dh-environment-setup`** | [`dh-environment-setup`](.agents/skills/dh-environment-setup/SKILL.md) | Registering Core endpoints (`dhcli register`), logging in (`dhcli login`), switching environments (`dhcli use`), bootstrapping `.venv` (`dhcli init`), or creating a Project. |
| **`dh-secure-and-configure-secrets`** | [`dh-secure-and-configure-secrets`](.agents/skills/dh-secure-and-configure-secrets/SKILL.md) | Creating project secrets via the Python SDK fallback and securely injecting credentials into runs via `spec.secrets`. |
| **`dh-data-and-storage`** | [`dh-data-and-storage`](.agents/skills/dh-data-and-storage/SKILL.md) | Uploading local files/datasets/models to S3 (`dhcli upload`), registering remote HTTPS/SQL tables, or downloading entities (`dhcli download`). |
| **`dh-author-and-run-python`** | [`dh-author-and-run-python`](.agents/skills/dh-author-and-run-python/SKILL.md) | Writing Python functions (`job`/`serve`), Base64-encoding handlers, applying image-reuse fast retries, and running via `dhcli run`. |
| **`dh-transform-with-dbt`** | [`dh-transform-with-dbt`](.agents/skills/dh-transform-with-dbt/SKILL.md) | Authoring SQL transformations in dbt, Base64-encoding `.sql` into a `kind: dbt` Function, and executing `dbt+transform` over `table` Dataitems. |
| **`dh-deploy-container-service`** | [`dh-deploy-container-service`](.agents/skills/dh-deploy-container-service/SKILL.md) | Running custom Docker/OCI containers as Kubernetes batch jobs (`container+job`), HTTP microservices/UIs (`container+serve`), or deployments (`container+deploy`). |
| **`dh-serve-ml-and-llm`** | [`dh-serve-ml-and-llm`](.agents/skills/dh-serve-ml-and-llm/SKILL.md) | Deploying Scikit-Learn (`sklearnserve`), MLflow (`mlflowserve`), HuggingFace (`huggingfaceserve`), vLLM (`vllmserve-*`), or KubeAI (`kubeai-*`) services. |
| **`dh-guardrail-and-evaluate-llm`** | [`dh-guardrail-and-evaluate-llm`](.agents/skills/dh-guardrail-and-evaluate-llm/SKILL.md) | Deploying real-time LLM input/output safety guardrails (`guardrail+serve`) and standardized KServe OpenInference v2 endpoints (`openinference+serve`). |
| **`dh-orchestrate-workflow`** | [`dh-orchestrate-workflow`](.agents/skills/dh-orchestrate-workflow/SKILL.md) | Authoring multi-step Hera (Argo) or KFP DAG pipelines in Python, registering via `dhcli create workflows`, executing `build` + `pipeline`, and attaching Triggers. |
| **`dh-troubleshoot-run`** | [`dh-troubleshoot-run`](.agents/skills/dh-troubleshoot-run/SKILL.md) | Diagnosing `ERROR` runs, retrieving logs via live container or persisted log fallback, and recovering from failures. |

---

## 3. Versioned Knowledge Base Map (`<%DOCS%>/knowledge/`)

### Pillar 1: Core Entities (`<%DOCS%>/knowledge/entities/`)
| Entity Page | Kinds & Scope |
| :--- | :--- |
| [`entities/index.md`](<%DOCS%>/knowledge/entities/index.md) | Master matrix of all 10 DigitalHub entities, `store://` URI formats, versioning, and `dhcli` CRUD support. |
| [`entities/project.md`](<%DOCS%>/knowledge/entities/project.md) | `project` — Top-level workspace, default file store config, and `dhcli create/get/list/delete projects`. |
| [`entities/function.md`](<%DOCS%>/knowledge/entities/function.md) | `function` — Executable template across all 16 runtime kinds, versioning, and `dhcli run` binding. |
| [`entities/workflow.md`](<%DOCS%>/knowledge/entities/workflow.md) | `hera`, `kfp` — Multi-step DAG definitions and two-phase `build` + `pipeline` execution. |
| [`entities/task.md`](<%DOCS%>/knowledge/entities/task.md) | `<kind>+<action>` — Intermediary task descriptor auto-resolved and created by `dhcli run`. |
| [`entities/run.md`](<%DOCS%>/knowledge/entities/run.md) | `<task>:run` — Execution instance, `run.yaml` `spec:` anatomy, state machine, `dhcli log`, `metrics`, `stop`. |
| [`entities/artifact.md`](<%DOCS%>/knowledge/entities/artifact.md) | `artifact` — Unstructured files/archives in S3 via `dhcli upload artifact` and `dhcli download artifact`. |
| [`entities/dataitem.md`](<%DOCS%>/knowledge/entities/dataitem.md) | `dataitem`, `table`, `croissant` — Tabular and structured datasets in S3 (`s3://`) or SQL (`sql://`). |
| [`entities/model.md`](<%DOCS%>/knowledge/entities/model.md) | `model`, `sklearn`, `mlflow`, `huggingface` — Versioned ML/DL model artifacts and serving lineage. |
| [`entities/secret.md`](<%DOCS%>/knowledge/entities/secret.md) | `secret` — Kubernetes-backed secrets mounted into runs via `spec.secrets`. |
| [`entities/trigger.md`](<%DOCS%>/knowledge/entities/trigger.md) | `scheduler`, `lifecycle` — Cron and event-driven run triggers. |

### Pillar 2: Execution Runtimes (`<%DOCS%>/knowledge/runtimes/`)
| Runtime Page | Function/Workflow Kinds | `dhcli run <task>` Actions |
| :--- | :--- | :--- |
| [`runtimes/index.md`](<%DOCS%>/knowledge/runtimes/index.md) | All 13 runtime families | Master runtime decision guide and `<kind>+<action>` task matrix. |
| [`runtimes/python.md`](<%DOCS%>/knowledge/runtimes/python.md) | `python` | `python+build`, `python+job`, `python+serve` |
| [`runtimes/container.md`](<%DOCS%>/knowledge/runtimes/container.md) | `container` | `container+build`, `container+job`, `container+serve`, `container+deploy` |
| [`runtimes/dbt.md`](<%DOCS%>/knowledge/runtimes/dbt.md) | `dbt` | `dbt+transform` |
| [`runtimes/hera.md`](<%DOCS%>/knowledge/runtimes/hera.md) | `hera` (workflow) | `hera+build`, `hera+pipeline` |
| [`runtimes/kfp.md`](<%DOCS%>/knowledge/runtimes/kfp.md) | `kfp` (workflow) | `kfp+build`, `kfp+pipeline` |
| [`runtimes/sklearnserve.md`](<%DOCS%>/knowledge/runtimes/sklearnserve.md) | `sklearnserve` | `sklearnserve+serve` |
| [`runtimes/mlflowserve.md`](<%DOCS%>/knowledge/runtimes/mlflowserve.md) | `mlflowserve` | `mlflowserve+build`, `mlflowserve+serve` |
| [`runtimes/huggingfaceserve.md`](<%DOCS%>/knowledge/runtimes/huggingfaceserve.md) | `huggingfaceserve` | `huggingfaceserve+serve` |
| [`runtimes/vllmserve.md`](<%DOCS%>/knowledge/runtimes/vllmserve.md) | `vllmserve-text`, `vllmserve-speech`, `vllmserve-pooling` | `vllmserve-text+serve`, `vllmserve-speech+serve`, `vllmserve-pooling+serve` |
| [`runtimes/kubeai.md`](<%DOCS%>/knowledge/runtimes/kubeai.md) | `kubeai-text`, `kubeai-speech` | `kubeai-text+serve`, `kubeai-speech+serve` |
| [`runtimes/openinference.md`](<%DOCS%>/knowledge/runtimes/openinference.md) | `openinference` | `openinference+build`, `openinference+serve` |
| [`runtimes/guardrail.md`](<%DOCS%>/knowledge/runtimes/guardrail.md) | `guardrail` | `guardrail+build`, `guardrail+serve` |
| [`runtimes/flower.md`](<%DOCS%>/knowledge/runtimes/flower.md) | `flower-app`, `flower-server`, `flower-client` | `flower-app+train`, `flower-server+build`, `flower-server+deploy`, `flower-client+build`, `flower-client+deploy` |

### Pillar 3: CLI Reference, In-Container SDK & Management
| Document | Scope |
| :--- | :--- |
| [`cli/index.md`](<%DOCS%>/knowledge/cli/index.md) | Complete `dhcli` command cheatsheet, global flags, and resource identifiers. |
| [`cli/auth-and-env.md`](<%DOCS%>/knowledge/cli/auth-and-env.md) | `.dhcore.ini` resolution, `register`, `login --pat`, `refresh`, `use`, `list-env`, `config`, `credentials`, `init`. |
| [`cli/crud-and-yaml-anatomy.md`](<%DOCS%>/knowledge/cli/crud-and-yaml-anatomy.md) | Universal YAML manifest envelope and `create`, `get`, `list`, `update`, `delete`. |
| [`cli/execution-and-runs.md`](<%DOCS%>/knowledge/cli/execution-and-runs.md) | `dhcli run`, `run.yaml` structure, polling `status.state`, `dhcli log`, `dhcli metrics`, `dhcli stop`. |
| [`cli/storage-and-services.md`](<%DOCS%>/knowledge/cli/storage-and-services.md) | S3 `upload`/`download` and service networking (`services`, `port-forward`, `proxy`). |
| [`sdk/handler-authoring.md`](<%DOCS%>/knowledge/sdk/handler-authoring.md) | Writing in-container Python `job` and `serve` (`init_context` + `serve`) handlers. |
| [`sdk/data-models-and-metrics.md`](<%DOCS%>/knowledge/sdk/data-models-and-metrics.md) | Reading `Dataitem` inputs (`as_df`) and logging outputs (`log_dataitem`, `log_sklearn`, `log_mlflow`, `log_metric`). |
| [`sdk/workflow-pipelines-dsl.md`](<%DOCS%>/knowledge/sdk/workflow-pipelines-dsl.md) | Authoring Hera (`digitalhub_runtime_hera.dsl`) and KFP (`digitalhub_runtime_kfp.dsl`) pipelines. |
| [`sdk/sdk-fallback.md`](<%DOCS%>/knowledge/sdk/sdk-fallback.md) | One-line Python SDK fallback recipes for running workflows, creating secrets, and attaching triggers. |
| [`management/keys-and-lifecycle.md`](<%DOCS%>/knowledge/management/keys-and-lifecycle.md) | `store://` URI grammar, source code URIs, entity versioning, Run state machine, and Console UI links. |
| [`management/kubernetes-resources.md`](<%DOCS%>/knowledge/management/kubernetes-resources.md) | `CoreResource` (`cpu`, `mem`, `gpu`), profiles, `CoreVolume` mounts, `envs`, `secrets`, `service_ports`. |
| [`management/governance-and-security.md`](<%DOCS%>/knowledge/management/governance-and-security.md) | RFC-1123 naming rules, metadata labels, and secret hygiene. |
| [`management/troubleshooting.md`](<%DOCS%>/knowledge/management/troubleshooting.md) | 3-step run triage protocol and error resolution matrix. |

---

## 4. Pre-Flight Verification Checklist Before Running `dhcli`

- [ ] `project` and `name` obey RFC-1123 (`^[a-z0-9-]+$`, max 63 chars).
- [ ] For `python` / `dbt` / `hera` / `kfp` / `openinference` / `guardrail` / `flower-*`, source code is nested under `spec.source.base64` (single-line Base64) and `spec.source.handler` is a valid `snake_case` function name.
- [ ] When uploading a new entity via `dhcli upload`, the `<resource>` argument is **singular** (`artifact`, `dataitem`, or `model`).
- [ ] When running a function via `dhcli run <kind>+<action> -p <project> -n <fn-name> -f run.yaml`, `run.yaml` contains `spec:` with `inputs` (`store://...`), `parameters`, and `resources`—and does **not** hardcode `spec.function` or `spec.task`.
- [ ] All `spec.envs[].value` entries are quoted strings, and sensitive tokens are referenced via `spec.secrets` rather than plaintext env vars.
