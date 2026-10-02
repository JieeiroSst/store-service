{{- define "common.hpa" -}}
{{- $autoscaling := .svc.autoscaling | default dict -}}
{{- if $autoscaling.enabled }}
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "common.name" . }}-hpa
  labels:
    {{- include "common.labels" . | nindent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: {{ include "common.workloadKind" . }}
    name: {{ include "common.workloadName" . }}
  minReplicas: {{ $autoscaling.minReplicas | default 1 }}
  maxReplicas: {{ $autoscaling.maxReplicas | default 3 }}
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{ $autoscaling.targetCPUUtilizationPercentage | default 70 }}
    {{- with $autoscaling.targetMemoryUtilizationPercentage }}
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: {{ . }}
    {{- end }}
  behavior:
    {{- with $autoscaling.behavior }}
    {{- toYaml . | nindent 4 }}
    {{- else }}
    # Scale up fast, scale down slowly so a short lull doesn't drop capacity
    # right before the next burst.
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
        - type: Percent
          value: 100
          periodSeconds: 30
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 25
          periodSeconds: 60
    {{- end }}
{{- end -}}
{{- end -}}
