## Minimal Deployment: app + dlv sidecar
   This is the simplest working example of a pod that runs:
   • your Go app
   • a Delve sidecar that attaches to it
   • shared PID namespace
   • correct ptrace capability

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: demo-app
  template:
    metadata:
      labels:
        app: demo-app
    spec:
      shareProcessNamespace: true
      containers:
      - name: app
        image: kind.local/demo-app:latest
        command: ["/app"]
      - name: dlv
        image: kind.local/dlv-sidecar:latest
        securityContext:
          capabilities:
            add: ["SYS_PTRACE"]
        command: [
          "sh", "-c",
          "APP_PID=$(pgrep -xo app) && \
          dlv attach $APP_PID --headless --listen=:40000 --api-version=2 --accept-multiclient"
          ]
        ports:
        - containerPort: 40000
          name: debug
```
  Your operator will generate this layout automatically when debug mode is enabled.

## Minimal Delve sidecar image

Need Delve, a shell, and pgrep:

```dockerfile
FROM alpine:3.20
RUN apk add --no-cache bash curl procps
RUN curl -L https://github.com/go-delve/delve/releases/download/v1.26.0/dlv-linux-arm64.tar.gz \
| tar -xz -C /usr/local/bin
ENTRYPOINT ["dlv"]
```

This image runs Delve directly as a debug sidecar.

## ko commands for kind

Use two workflows:
- ko apply - for the operator (because it has YAML)
- ko build - for app and sidecar (because they have no YAML)

### Build and deploy the operator

```bash
export KO_DOCKER_REPO=kind.local/operator
ko apply -f config/
```
This builds the operator, pushes it into kind, rewrites YAML to the digest, and applies it.

### Build the app image

```bash
export KO_DOCKER_REPO=kind.local/demo-app:latest
ko build ./cmd/app
```

### Build the Delve sidecar image

```bash
export KO_DOCKER_REPO=kind.local/dlv-sidecar:latest
ko build ./cmd/dlv-sidecar
```
Your operator can now reference `kind.local/demo-app:latest` and `kind.local/dlv-sidecar:latest`. No digests needed. No cloud registry.

## ko - debug‑friendly binaries

Debug flags, add them in ko.yaml:

```yaml
builds:
- id: app
  dir: ./cmd/app
  flags:
    - -gcflags=all=-N -l
```

This ensures Delve‑friendly builds every time.
