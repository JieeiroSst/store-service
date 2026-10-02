{{- define "common.pdb" -}}
{{- $pdb := .svc.pdb | default dict -}}
{{- if $pdb.enabled }}
---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: {{ include "common.name" . }}-pdb
  labels:
    {{- include "common.labels" . | nindent 4 }}
spec:
  {{- if hasKey $pdb "minAvailable" }}
  minAvailable: {{ $pdb.minAvailable }}
  {{- else }}
  maxUnavailable: {{ $pdb.maxUnavailable | default 1 }}
  {{- end }}
  # Let node drains evict pods that never became ready instead of blocking.
  unhealthyPodEvictionPolicy: AlwaysAllow
  selector:
    matchLabels:
      {{- include "common.selectorLabels" . | nindent 6 }}
{{- end -}}
{{- end -}}
