---
name: dh-guardrail-and-evaluate-llm
description: Deploy real-time LLM input/output safety guardrail microservices (guardrail+serve) and custom KServe OpenInference v2 endpoints (openinference+serve) on DigitalHub via dhcli.
---

# Skill: Deploying LLM Safety Guardrails & OpenInference Microservices

Use this procedural workflow whenever deploying input/output content safety filters (PII redaction, toxic language detection, prompt injection defense) or standardized KServe OpenInference v2 Python inference handlers.

## Reference Knowledge (Pinned to Platform Version)
- **Guardrail Runtime Specification**: [guardrail.md]({{DOCS}}/knowledge/runtimes/guardrail.md)
- **OpenInference Runtime Specification**: [openinference.md]({{DOCS}}/knowledge/runtimes/openinference.md)
- **Service Networking & Port-Forwarding**: [storage-and-services.md]({{DOCS}}/knowledge/cli/storage-and-services.md)
- **Kubernetes Compute Resources**: [kubernetes-resources.md]({{DOCS}}/knowledge/management/kubernetes-resources.md)

---

## 1. Pre-Flight Checklist
- [ ] Target `<project>` exists (`dhcli get projects <project> -o json`).
- [ ] Entity names follow RFC-1123 (`^[a-z0-9-]+$`).
- [ ] Python handler function name uses `snake_case`.

---

## 2. Step-by-Step Procedure

### Step 1: Write and Base64-Encode the Guardrail Script (`guard.py`)
```python
def validate_prompt(context, event):
    """
    Called on incoming requests to validate text against safety rules.
    Returns: {"allowed": bool, "sanitized_text": str, "reasons": list}
    """
    payload = event.body
    text = payload.get("text", "")

    # Basic heuristic check (replace with ML or regex filter)
    forbidden = ["DROP TABLE", "IGNORE PREVIOUS INSTRUCTIONS"]
    for phrase in forbidden:
        if phrase in text.upper():
            return {"allowed": False, "reason": f"Violates safety policy: {phrase}"}

    return {"allowed": True, "sanitized_text": text}
```

Encode the script:
```bash
base64 -i guard.py | tr -d '\n'
```

### Step 2: Create the `kind: guardrail` (or `kind: openinference`) Function
```yaml
kind: guardrail
name: prompt-guard-service
project: <project>
metadata:
  description: "Prompt injection and safety guardrail service"
spec:
  python_version: PYTHON3_10
  source:
    lang: python
    handler: validate_prompt
    base64: "<BASE64_ENCODED_SCRIPT>"
```
Apply via `dhcli`:
```bash
dhcli create functions -p <project> -f guardrail-fn.yaml
```

### Step 3: Launch the Service (`guardrail+serve` or `openinference+serve`)
Create `serve-run.yaml` configuring service ports and replicas:
```yaml
spec:
  replicas: 1
  service_type: ClusterIP
  service_ports:
    - port: 8080
      target_port: 8080
  resources:
    cpu: "1"
    mem: "2Gi"
```
Execute via `dhcli`:
```bash
dhcli run guardrail+serve -p <project> -n prompt-guard-service -f serve-run.yaml
```

### Step 4: Verification & Testing
1. Poll run state until `RUNNING`:
   ```bash
   dhcli get runs <run-id> -p <project> -o json | jq '{state:.status.state,service:.status.service}'
   ```
2. Retrieve internal cluster service URL:
   ```bash
   dhcli services -p <project> -o json
   ```
3. Test locally via port-forward tunnel:
   ```bash
   dhcli port-forward -p <project> -f prompt-guard-service -l 8080
   ```
   Send sample query in another terminal:
   ```bash
   curl -X POST http://localhost:8080/ -H "Content-Type: application/json" -d '{"text": "Hello world"}'
   ```
