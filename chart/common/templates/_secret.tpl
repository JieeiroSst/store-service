{{/*
Chart-owned Secrets:
  - envFile: the .env file a service loads at startup (godotenv), mounted at
    envFile.mountPath. With envFile.externalSecret it is synced from the secret
    manager instead of rendered from values.
  - secret: plain key/values for local and non-prod clusters only. Prod sets
    secret.create=false and syncs the same Secret name with externalSecret.
*/}}
{{- define "common.secret" -}}
{{- $svc := .svc -}}
{{- $envFile := $svc.envFile | default dict -}}
{{- $secret := $svc.secret | default dict -}}
{{- if and $envFile.enabled (not $envFile.externalSecret) }}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "common.name" . }}-env-file
  labels:
    {{- include "common.labels" . | nindent 4 }}
  annotations:
    argocd.argoproj.io/sync-wave: "-1"
type: Opaque
stringData:
  .env: |
    {{- tpl $envFile.content .root | nindent 4 }}
{{- end }}
{{- if and $envFile.enabled $envFile.externalSecret }}
{{- $es := $envFile.externalSecret }}
---
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: {{ include "common.name" . }}-env-file
  labels:
    {{- include "common.labels" . | nindent 4 }}
  annotations:
    argocd.argoproj.io/sync-wave: "-1"
spec:
  refreshInterval: {{ $es.refreshInterval | default "1h" }}
  secretStoreRef:
    kind: {{ ($es.secretStoreRef | default dict).kind | default "SecretStore" }}
    name: {{ ($es.secretStoreRef | default dict).name | default "store-secrets" }}
  target:
    name: {{ include "common.name" . }}-env-file
    creationPolicy: Owner
  data:
    - secretKey: .env
      remoteRef:
        key: {{ required "envFile.externalSecret.remoteKey is required" $es.remoteKey }}
        property: {{ $es.property | default "env" }}
{{- end }}
{{- if $secret.create }}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{ $secret.name | default (printf "%s-secret" (include "common.name" .)) }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
  annotations:
    argocd.argoproj.io/sync-wave: "-1"
type: Opaque
stringData:
  {{- range $k, $v := $secret.stringData }}
  {{ $k }}: {{ tpl (toString $v) $.root | toJson }}
  {{- end }}
{{- end }}
{{- end -}}
