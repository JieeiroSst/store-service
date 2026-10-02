{{/*
Seeds the service's Consul KV key from the chart's files/consul.json, only if
the key does not exist yet (hand-edited or rotated config is never
overwritten). Values:

  consulSeed:
    enabled: true
    key: vending_machine_service        # KV key the service reads (KeyConsul)
    file: files/consul.json             # default
    addr: http://consul-service         # default

Ordering differs per deployer:
  - Argo CD: a Sync hook in wave -1, so the key exists before the Deployment
    (wave 0) starts. A PostSync hook would never run while the app is still
    crash-looping for lack of config.
  - helm/umbrella: post-install/post-upgrade, because Consul itself is part of
    the same umbrella release.
*/}}
{{- define "common.consulSeed" -}}
{{- $seed := .svc.consulSeed | default dict -}}
{{- if $seed.enabled }}
{{- $name := printf "%s-consul-seed" (include "common.name" .) }}
{{- $key := required "consulSeed.key is required" $seed.key }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ $name }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
  annotations:
    argocd.argoproj.io/sync-wave: "-2"
data:
  {{ $key }}.json: |-
    {{- .root.Files.Get ($seed.file | default "files/consul.json") | nindent 4 }}
---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ $name }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
  annotations:
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/hook-weight": "1"
    "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
    argocd.argoproj.io/hook: Sync
    argocd.argoproj.io/sync-wave: "-1"
    argocd.argoproj.io/hook-delete-policy: BeforeHookCreation
spec:
  backoffLimit: 6
  activeDeadlineSeconds: 600
  ttlSecondsAfterFinished: 3600
  template:
    metadata:
      labels:
        app: {{ $name }}
        network.store.io/consul-client: "true"
    spec:
      restartPolicy: Never
      automountServiceAccountToken: false
      securityContext:
        runAsNonRoot: true
        runAsUser: 100
        runAsGroup: 101
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: consul-kv-seed
          # curl ships in the image: no package install, so no egress beyond Consul.
          image: {{ $seed.image | default "curlimages/curl:8.10.1" }}
          env:
            - name: CONSUL_ADDR
              value: {{ $seed.addr | default "http://consul-service" | quote }}
            - name: CONSUL_KEY
              value: {{ $key | quote }}
          command:
            - sh
            - -c
            - |
              set -e
              until curl -sf "$CONSUL_ADDR/v1/status/leader" | grep -q ':'; do
                echo "waiting for consul..."
                sleep 2
              done
              if curl -sf "$CONSUL_ADDR/v1/kv/$CONSUL_KEY?raw" >/dev/null; then
                echo "consul key $CONSUL_KEY already present, leaving as-is"
                exit 0
              fi
              curl -sf -X PUT --data-binary "@/seed/$CONSUL_KEY.json" "$CONSUL_ADDR/v1/kv/$CONSUL_KEY" >/dev/null
              echo "seeded consul key $CONSUL_KEY"
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: ["ALL"]
          resources:
            requests: { cpu: 10m, memory: 16Mi }
            limits: { cpu: 100m, memory: 64Mi }
          volumeMounts:
            - name: seed
              mountPath: /seed
      volumes:
        - name: seed
          configMap:
            name: {{ $name }}
{{- end }}
{{- end -}}
