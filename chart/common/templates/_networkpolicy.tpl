{{/*
Per-service allow rules. allowFromApps / allowToApps entries are
{name, namespace?, port} and match pods by their `app` label; namespace is only
needed when the peer runs outside the release namespace. Baseline rules shared by every service (DNS, Consul,
ingress controller, Prometheus) live in chart/platform and match on the
network.store.io/* pod labels, so only service-specific traffic goes here.
*/}}
{{- define "common.networkpolicy" -}}
{{- $np := .svc.networkPolicy | default dict -}}
{{- if $np.enabled }}
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: {{ include "common.name" . }}-netpol
  labels:
    {{- include "common.labels" . | nindent 4 }}
spec:
  podSelector:
    matchLabels:
      {{- include "common.selectorLabels" . | nindent 6 }}
  policyTypes:
    - Ingress
    - Egress
  {{- if not (or $np.allowFromApps $np.ingress) }}
  # Nothing beyond the platform baseline (ingress controller, Prometheus).
  ingress: []
  {{- else }}
  ingress:
    {{- range $np.allowFromApps }}
    - from:
        - podSelector:
            matchLabels:
              app: {{ .name }}
          {{- with .namespace }}
          namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: {{ . }}
          {{- end }}
      ports:
        - port: http
          protocol: TCP
    {{- end }}
    {{- with $np.ingress }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
  {{- end }}
  egress:
    {{- range $np.allowToApps }}
    - to:
        - podSelector:
            matchLabels:
              app: {{ .name }}
          {{- with .namespace }}
          namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: {{ . }}
          {{- end }}
      ports:
        - port: {{ .port }}
          protocol: TCP
    {{- end }}
    {{- with $np.egress }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
{{- end -}}
{{- end -}}
