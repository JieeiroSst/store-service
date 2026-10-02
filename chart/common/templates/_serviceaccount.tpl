{{- define "common.serviceaccount" -}}
{{- $sa := .svc.serviceAccount | default dict -}}
{{- if $sa.create }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "common.serviceAccountName" . }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
  annotations:
    argocd.argoproj.io/sync-wave: "-1"
    {{- with $sa.annotations }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
automountServiceAccountToken: {{ $sa.automountToken | default false }}
{{- end -}}
{{- end -}}
