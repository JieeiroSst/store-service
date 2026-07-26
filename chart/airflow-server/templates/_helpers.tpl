{{- define "airflow-server.sqlAlchemyConn" -}}
postgresql+psycopg2://{{ .Values.airflowServer.postgres.user }}:{{ .Values.airflowServer.postgres.password }}@{{ .Values.airflowServer.postgres.service.name }}:{{ .Values.airflowServer.postgres.port }}/{{ .Values.airflowServer.postgres.db }}
{{- end -}}

{{/*
Env vars shared by the webserver and scheduler containers so both processes
agree on the executor, metadata DB and API auth backend.
*/}}
{{- define "airflow-server.env" -}}
- name: AIRFLOW__CORE__EXECUTOR
  value: {{ .Values.airflowServer.executor | quote }}
- name: AIRFLOW__DATABASE__SQL_ALCHEMY_CONN
  value: {{ include "airflow-server.sqlAlchemyConn" . | quote }}
- name: AIRFLOW__API__AUTH_BACKENDS
  value: "airflow.api.auth.backend.basic_auth"
{{- if .Values.airflowServer.fernetKey }}
- name: AIRFLOW__CORE__FERNET_KEY
  value: {{ .Values.airflowServer.fernetKey | quote }}
{{- end }}
{{- end -}}
