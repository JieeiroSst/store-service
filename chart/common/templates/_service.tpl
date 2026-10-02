{{- define "common.service" -}}
{{- $svc := .svc -}}
{{- $ctx := . -}}
{{- $service := $svc.service | default dict -}}
{{- $metrics := $svc.metrics | default dict }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ include "common.serviceName" . }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
  {{- with $service.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  type: {{ $service.type | default "ClusterIP" }}
  selector:
    {{- include "common.selectorLabels" . | nindent 4 }}
  ports:
    - name: http
      port: {{ include "common.servicePort" . }}
      targetPort: http
      protocol: TCP
    {{- range $service.extraPorts }}
    - name: {{ .name }}
      port: {{ .port }}
      targetPort: {{ .targetPort | default .name }}
      protocol: {{ .protocol | default "TCP" }}
    {{- end }}
    {{- if and $metrics.enabled $metrics.port (ne (toString $metrics.port) (toString $svc.image.port)) }}
    - name: metrics
      port: {{ $metrics.port }}
      targetPort: metrics
      protocol: TCP
    {{- end }}
{{- /* Additional Services on the same pods, e.g. a separate gRPC endpoint. */}}
{{- range $svc.extraServices }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .name }}
  labels:
    {{- include "common.labels" $ctx | nindent 4 }}
  {{- with .annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  type: {{ .type | default "ClusterIP" }}
  selector:
    {{- include "common.selectorLabels" $ctx | nindent 4 }}
  ports:
    {{- range .ports }}
    - name: {{ .name }}
      port: {{ .port }}
      targetPort: {{ .targetPort | default .name }}
      protocol: {{ .protocol | default "TCP" }}
    {{- end }}
{{- end }}
{{- end -}}
