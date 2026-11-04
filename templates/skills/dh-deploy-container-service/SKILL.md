---
name: dh-deploy-container-service
description: Register custom Docker/OCI container Functions and run them as Kubernetes batch jobs (container+job), HTTP microservices/UIs (container+serve), persistent deployments (container+deploy), or Kaniko image builds (container+build) via dhcli.
---

# Skill: Deploying Custom Containers & HTTP Microservices

Use this procedural workflow whenever running a pre-built Docker/OCI image as a batch job (`container+job`), deploying a web UI or REST API (`container+serve`, such as Streamlit, Gradio, FastAPI), running a background worker (`container+deploy`), or building a container image in-cluster (`container+build`).

## Reference Knowledge (Pinned to Platform Version)
- **Container Runtime Specification**: [container.md](<%DOCS%>/knowledge/runtimes/container.md)
- **Kubernetes Resources, Volumes, Secrets & Ports**: [kubernetes-resources.md](<%DOCS%>/knowledge/management/kubernetes-resources.md)
- **Service Discovery & Port-Forwarding**: [storage-and-services.md](<%DOCS%>/knowledge/cli/storage-and-services.md)

---

## 1. Pre-Flight Checklist
- [ ] Target `<project>` exists (`dhcli get projects <project> -o json`).
- [ ] Function name obeys RFC-1123 (`^[a-z0-9-]+$`).
- [ ] Container image is publicly accessible or cluster has registry pull secrets.

---

## 2. Step-by-Step Procedure

### Step 1: Create the `kind: container` Function (`function.yaml`)
Define the container image, default entrypoint `command`/`args`, and image pull policy:
```yaml
kind: container
name: <function-name>
project: <project>
metadata:
  description: "<description>"
spec:
  image: "<registry>/<org>/<image>:<tag>"
  image_pull_policy: Always
  command: ["<entrypoint-binary>"]
  args: ["<default-arg-1>"]
```
```bash
dhcli create functions -p <project> -f function.yaml
dhcli get functions -p <project> -n <function-name> -o json
```

### Step 2A: Run as a Batch Job (`container+job`)
Create `job-run.yaml` (only `spec:` is needed; `dhcli run` populates `spec.function` and `spec.task` automatically):
```yaml
spec:
  args: ["--input", "/mnt/data/in.parquet"]
  envs:
    - name: LOG_LEVEL
      value: "INFO"
  secrets:
    - MY_SECRET_KEY
  resources:
    cpu: "2"
    mem: "4Gi"
```
```bash
dhcli run container+job -p <project> -n <function-name> -f job-run.yaml
```

### Step 2B: Deploy as a Long-Lived HTTP Service (`container+serve`)
Create `serve-run.yaml` specifying `service_ports` (`port` is the K8s Service port; `target_port` is the container listening port):
```yaml
spec:
  replicas: 1
  service_type: ClusterIP
  service_ports:
    - port: 80
      target_port: 8000
  envs:
    - name: PORT
      value: "8000"
  secrets:
    - API_SECRET
  resources:
    cpu: "1"
    mem: "2Gi"
```
```bash
dhcli run container+serve -p <project> -n <function-name> -f serve-run.yaml
```

### Step 3: Verify Run State, Service URL & Local Port-Forwarding
1. Poll run state until `RUNNING` (for services) or `COMPLETED` (for jobs):
   ```bash
   dhcli get runs <run-id> -p <project> -o json | jq '{state:.status.state,service:.status.service}'
   ```
2. For `container+serve` runs, inspect active service URLs across the project:
   ```bash
   dhcli services -p <project> -o json
   ```
3. Forward a local port to test the service from your workstation:
   ```bash
   dhcli port-forward -p <project> -f <function-name> -l 8000
   ```
