## Step 1 — Define CRD types in podsvc_types.go
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