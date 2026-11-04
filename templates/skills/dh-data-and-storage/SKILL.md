---
name: dh-data-and-storage
description: Upload local files, datasets, and model weights to S3 via dhcli upload, register external S3, HTTPS, or SQL tables via dhcli create, and download entities locally via dhcli download.
---

# Skill: Managing Dataitems, Artifacts & Models in Storage

Use this procedural workflow whenever uploading local files/directories to the project's S3 datalake, registering structured `table` Dataitems (`s3://`, `https://`, or `sql://`), registering public artifacts, or downloading entities to local disk.

## Reference Knowledge (Pinned to Platform Version)
- **Dataitem Entity (`dataitem`, `table`, `croissant`)**: [dataitem.md](<%DOCS%>/knowledge/entities/dataitem.md)
- **Artifact Entity (`artifact`)**: [artifact.md](<%DOCS%>/knowledge/entities/artifact.md)
- **Model Entity (`model`, `sklearn`, `mlflow`, `huggingface`)**: [model.md](<%DOCS%>/knowledge/entities/model.md)
- **CLI Storage & Services Reference**: [storage-and-services.md](<%DOCS%>/knowledge/cli/storage-and-services.md)

---

## 1. Pre-Flight Checklist
- [ ] Target `<project>` exists (`dhcli get projects <project> -o json`).
- [ ] Entity names obey RFC-1123 (`^[a-z0-9-]+$`, e.g. `raw-traffic-data`, not `raw_traffic.csv`).

---

## 2. Step-by-Step Procedure

### Scenario A: Upload a Local File or Directory to S3 (`dhcli upload`)
`dhcli upload` generates a new version UUID, streams the local file or folder to `s3://datalake/<project>/<resource>/<name>/<uuid>/...`, and transitions the entity state from `CREATED` to `READY`.

> **CRITICAL RULE**: Always pass the **singular** resource name (`artifact`, `dataitem`, or `model`) to `dhcli upload` when creating a new entity, because `dhcli upload` sets the entity `kind` equal to the `<resource>` argument.

```bash
# Upload an unstructured file or archive as kind: artifact
dhcli upload artifact -p <project> -n <artifact-name> -f ./path/to/file.tar.gz -v

# Upload a CSV or Parquet file as kind: dataitem
dhcli upload dataitem -p <project> -n <dataitem-name> -f ./path/to/data.parquet -v

# Upload generic model weights or directory as kind: model
dhcli upload model -p <project> -n <model-name> -f ./path/to/model_dir -v
```

Verify the uploaded entity and retrieve its `<uuid>` and `store://` key (`store://<project>/<entity-type>/<kind>/<name>:<uuid>`):
```bash
dhcli get dataitems -p <project> -n <dataitem-name> -o json
```

### Scenario B: Upload a Local File into a Typed Kind (`table`, `sklearn`, `mlflow`, `huggingface`)
When you need a specific sub-kind such as `kind: table` (for Dataitems) or `kind: sklearn` / `mlflow` / `huggingface` (for Models):
1. Create the typed metadata manifest first via `dhcli create`:
   ```yaml
   # table.yaml
   kind: table
   name: <dataset-name>
   project: <project>
   metadata:
     description: "Tabular dataset stored in S3"
   spec:
     path: "s3://datalake/<project>/dataitem/<dataset-name>/data.parquet"
   ```
   ```bash
   dhcli create dataitems -p <project> -f table.yaml
   ```
2. Fetch the newly created `<uuid>`:
   ```bash
   dhcli get dataitems -p <project> -n <dataset-name> -o json
   ```
3. Upload the local file into that specific `<uuid>`:
   ```bash
   dhcli upload dataitem <uuid> -p <project> -f ./data.parquet -v
   ```

### Scenario C: Register an Existing External `s3://`, `http(s)://`, or `sql://` Dataset or Artifact (`dhcli create`)
If the dataset or file already exists in remote storage or on the web, register its catalog metadata directly with `dhcli create` (no upload needed):
- For a remote table (`s3://...`, `https://.../data.csv`, or `sql://...`):
  ```yaml
  kind: table
  name: <table-name>
  project: <project>
  spec:
    path: "https://raw.githubusercontent.com/.../data.csv"
  ```
  ```bash
  dhcli create dataitems -p <project> -f external-table.yaml
  ```
- For a remote public artifact (e.g. PDF, audio, zip archive):
  ```yaml
  kind: artifact
  name: <artifact-name>
  project: <project>
  spec:
    path: "https://example.com/document.pdf"
  ```
  ```bash
  dhcli create artifacts -p <project> -f external-artifact.yaml
  ```
Inside a Python handler, `artifact.download()` and `table.as_df()` or `table.download()` resolve both S3 and `https://` URLs automatically.

### Scenario D: Download an Entity from S3 to Local Disk (`dhcli download`)
```bash
# Download the latest version by name (-n)
dhcli download dataitem -p <project> -n <dataitem-name> -d ./local_dir

# Download a specific version by UUID (<id>)
dhcli download model <uuid> -p <project> -d ./local_model
```
