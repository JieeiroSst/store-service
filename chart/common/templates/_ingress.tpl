{{- define "common.ingress" -}}
{{- $svc := .svc -}}
{{- $ingress := $svc.ingress | default dict -}}
{{- if $ingress.enabled }}
{{- include "common.ingressResource" (dict "ctx" . "ingress" $ingress "name" (printf "%s-ingress" (include "common.name" .)) "serviceName" (include "common.serviceName" .) "servicePort" "http") }}
{{- end }}
{{- /* Additional Ingresses, e.g. a gRPC host routed to an extraServices entry. */}}
{{- range $svc.extraIngresses }}
{{- if .enabled }}
{{- include "common.ingressResource" (dict "ctx" $ "ingress" . "name" .name "serviceName" .serviceName "servicePort" .servicePort) }}
{{- end }}
{{- end }}
{{- end -}}

{{- define "common.ingressResource" -}}
{{- $ingress := .ingress -}}
{{- $tls := $ingress.tls | default dict }}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ .name }}
  labels:
    {{- include "common.labels" .ctx | nindent 4 }}
  {{- with $ingress.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  ingressClassName: {{ $ingress.className | default "nginx" }}
  {{- if $tls.enabled }}
  tls:
    - hosts:
        - {{ $ingress.host | quote }}
      secretName: {{ $tls.secretName | default (printf "%s-tls" .name) }}
  {{- end }}
  rules:
    - host: {{ $ingress.host | quote }}
      http:
        paths:
          - path: {{ $ingress.path | default "/" }}
            pathType: {{ $ingress.pathType | default "Prefix" }}
            backend:
              service:
                name: {{ .serviceName }}
                port:
                  {{- if kindIs "string" .servicePort }}
                  name: {{ .servicePort }}
                  {{- else }}
                  number: {{ .servicePort }}
                  {{- end }}
{{- end -}}
