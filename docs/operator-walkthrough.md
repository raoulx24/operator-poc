## Step 0 - Scaffolding

```sh
operator-sdk init --domain=my.domain --owner="raoulx24"
go mod init github.com/raoulx24/operator-poc
operator-sdk create api --group=operatorpoc --version=v1alpha1 --kind=PodSvc --resource --controller
```

## Step 1 - Define CRD types in podsvc_types.go
We will:

- Define PodSvcSpec with:
  - LabelName (string, required)
  - LabelValue (string, required)
  - Ports []corev1.ServicePort (required, minItems=1)
- Define PodSvcStatus with:
  - Entries []PodSvcStatusEntry
- Define PodSvcStatusEntry with:
  - PodName
  - ServiceName
  - MatchedPorts []corev1.ServicePort
  - UnmatchedPorts []UnmatchedPortStatus
- Add kubebuilder validation markers
- Add kubebuilder printcolumns (optional but nice)
- Regenerate CRDs afterward

## Step 2 - Regenerate CRDs and validate the schema

After defining the CRD types, regenerate all code and CRD manifests to ensure the YAML definitions match the Go structs. This step uses controller-gen via `make generate` and `make manifests`.

Verify that the generated CRD under `config/crd/bases/` includes the correct OpenAPI schema for:
- `spec.labelName`
- `spec.labelValue`
- `spec.ports`
- `status.entries[*]` with `podName`, `serviceName`, `matchedPorts`, `unmatchedPorts`

Validate the CRD structure using `kubectl apply --dry-run=client`

## Step 3 - Implement the controller logic

The controller is now responsible for reconciling PodSvc resources. The reconciliation loop performs the following:

1. Fetch the PodSvc instance.
2. List all Pods matching the label selector defined in the CR.
3. For each matching Pod:
    - Compute matched and unmatched ports by comparing CR ports with container ports.
    - Create or update a dedicated Service named `<podsvc>-<pod>`.
    - Record matched/unmatched ports in the PodSvc status.
4. Identify and delete orphaned Services that were previously created but no longer correspond to any matching Pod.
5. Update the PodSvc status with the full list of status entries.

This step establishes the core behavior of the operator: dynamically managing per-Pod Services and maintaining accurate status information.


## Step 4 — Deploy and test the operator

Deploy the operator into a Kubernetes cluster. Installed the CRD using `make install` and deploy the controller with `make deploy`. After confirming the controller pod was running, create a test Deployment with labeled Pods and applied a PodSvc resource pointing to those labels.

```yaml
apiVersion: /v1alpha1
kind: PodSvc
metadata:
  name: demo-svc
spec:
  labelName: app
  labelValue: demo
  ports:
  - name: http
    port: 80
    targetPort: 80
    protocol: TCP

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo
spec:
  replicas: 2
  selector:
    matchLabels:
      app: demo
  template:
    metadata:
      labels:
        app: demo
    spec:
      containers:
      - name: demo
        image: nginx
        ports:
        - containerPort: 80
```

## Note
Operator‑SDK v4 uses a different project layout than Kubebuilder v3.

Controllers are located under `internal/controller/`, not `controllers/`.

The controller is automatically registered through internal/controller/controller.go 
and invoked from cmd/main.go. The correct place to implement reconciliation logic 
is internal/controller/podsvc_controller.go.
