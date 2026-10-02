{{- define "airflow-server.sqlAlchemyConn" -}}
postgresql+psycopg2://{{ .Values.airflowServer.postgres.user }}:{{ .Values.airflowServer.postgres.password }}@{{ .Values.airflowServer.postgres.service.name }}:{{ .Values.airflowServer.postgres.service.port }}/{{ .Values.airflowServer.postgres.db }}
{{- end -}}

{{- define "airflow-server.secretName" -}}
{{ .Values.airflowServer.name }}-secrets
{{- end -}}

{{/*
A component (webserver / scheduler) as a chart/common service block: the
shared Airflow image and the env both processes must agree on (executor,
metadata DB, API auth backend, fernet key) are filled in here so they are
declared once in values.yaml.
*/}}
{{- define "airflow-server.component" -}}
{{- $c := deepCopy .component -}}
{{- $image := deepCopy .root.Values.airflowServer.image -}}
{{- with $c.port }}{{ $_ := set $image "port" . }}{{ end -}}
{{- $_ := set $c "image" $image -}}
{{- $_ := set $c "containerEnv" (concat .root.Values.airflowServer.sharedEnv ($c.containerEnv | default list)) -}}
{{- toYaml $c -}}
{{- end -}}
