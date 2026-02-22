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


