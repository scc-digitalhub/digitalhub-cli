---
name: dh-transform-with-dbt
description: Author SQL data transformations using dbt, package them into kind dbt Functions, and execute dbt+transform batch jobs over tabular Dataitems via dhcli.
---

# Skill: Data Transformation with dbt (`kind: dbt`)

Use this procedural workflow whenever running SQL data engineering transformations, aggregations, and cleaning pipelines over tabular Dataitems stored in PostgreSQL or Dremio using dbt.

## Reference Knowledge (Pinned to Platform Version)
- **DBT Runtime Specification**: [dbt.md](<%DOCS%>/knowledge/runtimes/dbt.md)
- **Dataitem Entity (`kind: table`)**: [dataitem.md](<%DOCS%>/knowledge/entities/dataitem.md)
- **CLI Execution & Runs**: [execution-and-runs.md](<%DOCS%>/knowledge/cli/execution-and-runs.md)

---

## 1. Pre-Flight Checklist
- [ ] Target `<project>` exists (`dhcli get projects <project> -o json`).
- [ ] Source table Dataitem exists and is in `READY` state (`dhcli get dataitems -p <project> -n <table-name> -o json`).
- [ ] Function name obeys RFC-1123 (`^[a-z0-9-]+$`).

---

## 2. Step-by-Step Procedure

### Step 1: Write and Base64-Encode the dbt SQL Model (`model.sql`)
Write your standard dbt SELECT transformation:
```sql
{{ config(materialized='table') }}

SELECT
    id,
    customer_id,
    amount,
    created_at
FROM {{ ref('raw_transactions') }}
WHERE amount > 0
```

Encode the `.sql` file:
```bash
base64 -i model.sql | tr -d '\n'
```

### Step 2: Create the `kind: dbt` Function Manifest
```yaml
kind: dbt
name: transform-transactions
project: <project>
metadata:
  description: "Clean transactions data transformation model"
spec:
  source:
    lang: sql
    handler: model
    base64: "<BASE64_ENCODED_SQL>"
```
Register the function:
```bash
dhcli create functions -p <project> -f dbt-function.yaml
```

### Step 3: Execute the Transformation (`dbt+transform`)
Create `transform-run.yaml` mapping inputs and outputs:
```yaml
spec:
  inputs:
    raw_transactions: "store://<project>/dataitem/table/raw_transactions:latest"
  resources:
    cpu: "1"
    mem: "2Gi"
```
Execute the transformation job:
```bash
dhcli run dbt+transform -p <project> -n transform-transactions -f transform-run.yaml
```

### Step 4: Verification & Log Inspection
1. Poll the run state until `COMPLETED`:
   ```bash
   dhcli get runs <run-id> -p <project> -o json | jq '{state:.status.state,outputs:.status.outputs}'
   ```
2. Verify container logs (live or persisted):
   ```bash
   dhcli log <run-id> -p <project>
   # If returns "container not found":
   dhcli list logs -p <project> -o json | jq -r '.[] | select(.run == "<run-id>") | .content | @base64d'
   ```
3. Inspect the newly transformed output table:
   ```bash
   dhcli list dataitems -p <project> -o json
   ```
