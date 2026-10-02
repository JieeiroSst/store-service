{{- define "common.deployment" -}}
{{- $svc := .svc -}}
{{- $root := .root -}}
{{- $autoscaling := $svc.autoscaling | default dict -}}
{{- $strategy := $svc.strategy | default dict -}}
{{- $probes := $svc.probes | default dict -}}
{{- $metrics := $svc.metrics | default dict -}}
{{- $topology := $svc.topologySpread | default dict -}}
{{- $envFile := $svc.envFile | default dict -}}
{{- $mountsTmp := false -}}
{{- range $svc.volumeMounts }}{{ if eq .mountPath "/tmp" }}{{ $mountsTmp = true }}{{ end }}{{ end }}
{{- $kind := include "common.workloadKind" . }}
---
apiVersion: apps/v1
kind: {{ $kind }}
metadata:
  name: {{ include "common.workloadName" . }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
spec:
  {{- if not $autoscaling.enabled }}
  replicas: {{ $svc.replicas | default 1 }}
  {{- end }}
  revisionHistoryLimit: {{ $svc.revisionHistoryLimit | default 5 }}
  {{- with $svc.progressDeadlineSeconds }}
  progressDeadlineSeconds: {{ . }}
  {{- end }}
  {{- if eq $kind "StatefulSet" }}
  # Stable per-pod DNS (<pod>.<serviceName>.<namespace>.svc).
  serviceName: {{ required "statefulSet.serviceName is required" ($svc.statefulSet | default dict).serviceName }}
  {{- with ($svc.statefulSet | default dict).podManagementPolicy }}
  podManagementPolicy: {{ . }}
  {{- end }}
  updateStrategy:
    type: RollingUpdate
  {{- else }}
  strategy:
    {{- if eq ($strategy.type | default "RollingUpdate") "Recreate" }}
    # ReadWriteOnce volume: the old pod must release it before the new one starts.
    type: Recreate
    {{- else }}
    type: RollingUpdate
    rollingUpdate:
      maxSurge: {{ $strategy.maxSurge | default "25%" }}
      maxUnavailable: {{ $strategy.maxUnavailable | default 0 }}
    {{- end }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "common.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "common.labels" . | nindent 8 }}
        {{- with include "common.networkLabels" . | trim }}
        {{- . | nindent 8 }}
        {{- end }}
        {{- with $svc.podLabels }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
      {{- if or $svc.podAnnotations $metrics.enabled $envFile.enabled }}
      annotations:
        {{- with $svc.podAnnotations }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
        {{- if $metrics.enabled }}
        prometheus.io/scrape: "true"
        prometheus.io/path: {{ $metrics.path | default "/metrics" | quote }}
        prometheus.io/port: {{ $metrics.port | default $svc.image.port | quote }}
        {{- end }}
        {{- if $envFile.enabled }}
        # Roll the pods when the mounted .env changes; the app only reads it at startup.
        checksum/env-file: {{ list (tpl ($envFile.content | default "") $root) $envFile.externalSecret | toYaml | sha256sum }}
        {{- end }}
      {{- end }}
    spec:
      serviceAccountName: {{ include "common.serviceAccountName" . }}
      automountServiceAccountToken: {{ ($svc.serviceAccount | default dict).automountToken | default false }}
      {{- with $svc.priorityClassName }}
      priorityClassName: {{ . }}
      {{- end }}
      {{- with $svc.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      terminationGracePeriodSeconds: {{ $svc.terminationGracePeriodSeconds | default 30 }}
      securityContext:
        {{- toYaml ($svc.podSecurityContext | default dict) | nindent 8 }}
      {{- if $topology.enabled }}
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: kubernetes.io/hostname
          whenUnsatisfiable: {{ $topology.hostnameWhenUnsatisfiable | default "ScheduleAnyway" }}
          labelSelector:
            matchLabels:
              {{- include "common.selectorLabels" . | nindent 14 }}
          matchLabelKeys:
            - pod-template-hash
        - maxSkew: 1
          topologyKey: topology.kubernetes.io/zone
          whenUnsatisfiable: {{ $topology.zoneWhenUnsatisfiable | default "ScheduleAnyway" }}
          labelSelector:
            matchLabels:
              {{- include "common.selectorLabels" . | nindent 14 }}
          matchLabelKeys:
            - pod-template-hash
      {{- end }}
      {{- with $svc.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $svc.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $svc.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- /* initContainersTemplate names a chart-local `define` so init containers
             can keep their own template logic (conditionals, value references). */}}
      {{- $init := "" }}
      {{- if $svc.initContainersTemplate }}
      {{- $init = include $svc.initContainersTemplate $root | trim }}
      {{- else if $svc.initContainers }}
      {{- $init = tpl (toYaml $svc.initContainers) $root | trim }}
      {{- end }}
      {{- with $init }}
      {{- /* Init containers inherit the app container's securityContext unless they
             set their own, so the pod as a whole meets the namespace's Pod Security level. */}}
      {{- $inits := fromYamlArray . }}
      {{- range $inits }}
      {{- if not (hasKey . "securityContext") }}
      {{- $_ := set . "securityContext" ($svc.securityContext | default dict) }}
      {{- end }}
      {{- end }}
      initContainers:
        {{- toYaml $inits | nindent 8 }}
      {{- end }}
      containers:
        {{- /* Sidecars share the pod (and its hardening) with the app container. */}}
        {{- $sidecarsYaml := "" }}
        {{- if $svc.sidecarsTemplate }}
        {{- $sidecarsYaml = include $svc.sidecarsTemplate $root | trim }}
        {{- else if $svc.sidecars }}
        {{- $sidecarsYaml = tpl (toYaml $svc.sidecars) $root }}
        {{- end }}
        {{- with $sidecarsYaml }}
        {{- $sidecars := fromYamlArray . }}
        {{- range $sidecars }}
        {{- if not (hasKey . "securityContext") }}
        {{- $_ := set . "securityContext" ($svc.securityContext | default dict) }}
        {{- end }}
        {{- end }}
        {{- toYaml $sidecars | nindent 8 }}
        {{- end }}
        - name: {{ $svc.containerName | default (include "common.name" .) }}
          image: {{ include "common.image" . | quote }}
          imagePullPolicy: {{ $svc.image.pullPolicy | default "IfNotPresent" }}
          {{- with $svc.command }}
          command:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          {{- with $svc.args }}
          args:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          {{- if or $svc.image.port $svc.extraPorts }}
          ports:
            {{- with $svc.image.port }}
            - name: http
              containerPort: {{ . }}
              protocol: TCP
            {{- end }}
            {{- range $svc.extraPorts }}
            - name: {{ .name }}
              containerPort: {{ .containerPort }}
              protocol: {{ .protocol | default "TCP" }}
            {{- end }}
            {{- if and $metrics.enabled $metrics.port (ne (toString $metrics.port) (toString $svc.image.port)) }}
            - name: metrics
              containerPort: {{ $metrics.port }}
              protocol: TCP
            {{- end }}
          {{- end }}
          {{- with $svc.containerEnv }}
          env:
            {{- include "common.env" (dict "root" $root "env" .) | nindent 12 }}
          {{- end }}
          {{- with $svc.envFrom }}
          envFrom:
            {{- tpl (toYaml .) $root | nindent 12 }}
          {{- end }}
          {{- with $probes.startup }}
          startupProbe:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          {{- with $probes.readiness }}
          readinessProbe:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          {{- with $probes.liveness }}
          livenessProbe:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          {{- if $svc.lifecycle }}
          lifecycle:
            {{- toYaml $svc.lifecycle | nindent 12 }}
          {{- else if $svc.preStopSleepSeconds }}
          lifecycle:
            # Give kube-proxy / ingress controllers time to drop the pod from
            # their endpoints before the process receives SIGTERM.
            preStop:
              sleep:
                seconds: {{ $svc.preStopSleepSeconds }}
          {{- end }}
          securityContext:
            {{- toYaml ($svc.securityContext | default dict) | nindent 12 }}
          resources:
            {{- toYaml ($svc.resources | default dict) | nindent 12 }}
          volumeMounts:
            {{- if not $mountsTmp }}
            # readOnlyRootFilesystem needs somewhere writable for temp files.
            - name: tmp
              mountPath: /tmp
            {{- end }}
            {{- if $envFile.enabled }}
            - name: env-file
              mountPath: {{ $envFile.mountPath | default "/app/.env" }}
              subPath: .env
              readOnly: true
            {{- end }}
            {{- with include "common.whenList" (dict "root" $root "list" $svc.volumeMounts) | trim }}
            {{- . | nindent 12 }}
            {{- end }}
      volumes:
        {{- if not $mountsTmp }}
        - name: tmp
          emptyDir:
            sizeLimit: {{ $svc.tmpSizeLimit | default "256Mi" }}
        {{- end }}
        {{- if $envFile.enabled }}
        - name: env-file
          secret:
            secretName: {{ include "common.name" . }}-env-file
        {{- end }}
        {{- with include "common.whenList" (dict "root" $root "list" $svc.volumes) | trim }}
        {{- . | nindent 8 }}
        {{- end }}
{{- end -}}

{{/*
Container env. Each entry is {name, value} or {name, valueFrom}; string values
go through `tpl`, so they may reference other values, e.g.
  - name: MYSQL_HOST
    value: "{{ .Values.paymentService.env.mysqlHost }}"
Order is kept, so $(VAR) references keep working. An entry with `when` is
only rendered if that expression renders non-empty (optional integrations):
  - name: TEMPORAL_ADDRESS
    value: "{{ .Values.x.env.temporalAddress }}"
    when: "{{ .Values.x.env.temporalAddress }}"
*/}}
{{- define "common.env" -}}
{{- range .env }}
{{- if or (not (hasKey . "when")) (tpl (toString .when) $.root) }}
- name: {{ .name }}
  {{- if hasKey . "valueFrom" }}
  valueFrom:
    {{- tpl (toYaml .valueFrom) $.root | nindent 4 }}
  {{- else }}
  value: {{ if kindIs "invalid" .value }}""{{ else }}{{ tpl (toString .value) $.root | toJson }}{{ end }}
  {{- end }}
{{- end }}
{{- end }}
{{- end -}}

{{/*
Renders a list of volumes / volumeMounts through `tpl`, skipping entries whose
`when` expression renders empty (same rule as common.env).
*/}}
{{- define "common.whenList" -}}
{{- range .list }}
{{- if or (not (hasKey . "when")) (tpl (toString .when) $.root) }}
{{ tpl (toYaml (list (omit . "when"))) $.root }}
{{- end }}
{{- end }}
{{- end -}}
