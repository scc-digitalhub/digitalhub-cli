---
name: dh-serve-ml-and-llm
description: Deploy machine learning models (sklearnserve, mlflowserve, huggingfaceserve) and generative LLMs/Speech services (vllmserve, kubeai) as autoscaled Kubernetes endpoints via dhcli.
---

# Skill: Serving Machine Learning Models & Large Language Models

Use this procedural workflow whenever deploying trained Classical ML models (`sklearn`, `mlflow`, `huggingface`) as KServe OpenInference v2 services or deploying Generative AI models (LLMs, Whisper speech-to-text, embedding models) via vLLM or KubeAI.

## Reference Knowledge (Pinned to Platform Version)
- **Serving Runtimes Overview**: [runtimes/index.md](<%DOCS%>/knowledge/runtimes/index.md)
- **KubeAI Runtime (`kubeai-text`, `kubeai-speech`)**: [kubeai.md](<%DOCS%>/knowledge/runtimes/kubeai.md)
- **vLLM Serving Runtime (`vllmserve-*`)**: [vllmserve.md](<%DOCS%>/knowledge/runtimes/vllmserve.md)
- **Classical Model Serving**: [sklearnserve.md](<%DOCS%>/knowledge/runtimes/sklearnserve.md), [mlflowserve.md](<%DOCS%>/knowledge/runtimes/mlflowserve.md), [huggingfaceserve.md](<%DOCS%>/knowledge/runtimes/huggingfaceserve.md)
- **Model Entity Specification**: [model.md](<%DOCS%>/knowledge/entities/model.md)

---

## 1. Pre-Flight Checklist
- [ ] Target `<project>` exists (`dhcli get projects <project> -o json`).
- [ ] Model entity is in `READY` state (`dhcli get models -p <project> -n <model-name> -o json`).
- [ ] Function name obeys RFC-1123 (`^[a-z0-9-]+$`).

---

## 2. Step-by-Step Procedure

### Track A: Serving Classical ML Models (`sklearnserve`, `mlflowserve`, `huggingfaceserve`)

1. **Create the Serving Function (`function.yaml`)**:
   ```yaml
   kind: sklearnserve
   name: fraud-detector-api
   project: <project>
   metadata:
     description: "Scikit-Learn fraud inference service"
   spec:
     model: "store://<project>/model/sklearn/fraud-rf:<uuid>"
   ```
   ```bash
   dhcli create functions -p <project> -f function.yaml
   ```

2. **Deploy the Inference Service (`<runtime>+serve`)**:
   ```yaml
   # serve-run.yaml
   spec:
     replicas: 1
     resources:
       cpu: "1"
       mem: "2Gi"
   ```
   ```bash
   dhcli run sklearnserve+serve -p <project> -n fraud-detector-api -f serve-run.yaml
   ```

3. **Verify & Query V2 OpenInference Endpoint**:
   Once `status.state` is `RUNNING`, port-forward and send a V2 inference request:
   ```bash
   dhcli port-forward -p <project> -f fraud-detector-api -l 8080
   ```
   ```bash
   curl -X POST http://localhost:8080/v2/models/fraud-detector-api/infer \
     -H "Content-Type: application/json" \
     -d '{"inputs": [{"name": "input-0", "shape": [1, 4], "datatype": "FP32", "data": [[5.1, 3.5, 1.4, 0.2]]}]}'
   ```

---

### Track B: Serving LLMs, Whisper & Embeddings via KubeAI (`kubeai-text`, `kubeai-speech`)

1. **Create the KubeAI Function Manifest**:
   ```yaml
   kind: kubeai-text
   name: llama3-service
   project: <project>
   metadata:
     description: "Llama-3 8B Instruct chat endpoint"
   spec:
     model_name: llama-3-8b
     url: "hf://meta-llama/Meta-Llama-3-8B-Instruct"
     engine: VLLM
     features:
       - TextGeneration
     args:
       - "--max-model-len=4096"
     secrets:
       - HF_TOKEN
   ```
   ```bash
   dhcli create functions -p <project> -f function.yaml
   ```

2. **Deploy the Service with Hardware Profile & Autoscaling**:
   ```yaml
   # serve-run.yaml
   spec:
     profile: "1xa100"
     scaling:
       min_replicas: 0
       max_replicas: 2
   ```
   ```bash
   dhcli run kubeai-text+serve -p <project> -n llama3-service -f serve-run.yaml
   ```

3. **Verify OpenAI-Compatible Endpoint**:
   Retrieve the service URL from `dhcli services -p <project> -o json` or forward locally:
   ```bash
   dhcli port-forward -p <project> -f llama3-service -l 8000
   ```
   ```bash
   curl http://localhost:8000/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{"model": "llama-3-8b", "messages": [{"role": "user", "content": "Hello!"}]}'
   ```
