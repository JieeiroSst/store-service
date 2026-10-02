{{/*
Only rendered when the Prometheus Operator CRDs exist, so the same values work
on clusters that scrape through pod annotations instead.
*/}}
{{- define "common.servicemonitor" -}}
{{- $metrics := .svc.metrics | default dict -}}
{{- $sm := $metrics.serviceMonitor | default dict -}}
{{- if and $metrics.enabled $sm.enabled (.root.Capabilities.APIVersions.Has "monitoring.coreos.com/v1/ServiceMonitor") }}
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: {{ include "common.name" . }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
    {{- with $sm.labels }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
spec:
  selector:
    matchLabels:
      {{- include "common.selectorLabels" . | nindent 6 }}
  endpoints:
    - port: {{ if and $metrics.port (ne (toString $metrics.port) (toString .svc.image.port)) }}metrics{{ else }}http{{ end }}
      path: {{ $metrics.path | default "/metrics" }}
      interval: {{ $sm.interval | default "30s" }}
      scrapeTimeout: {{ $sm.scrapeTimeout | default "10s" }}
{{- end -}}
{{- end -}}
