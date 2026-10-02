{{- define "bonuslink-service.initContainers" -}}
{{- if .Values.bonuslinkService.migrations.enabled }}
- name: create-database
  image: {{ .Values.bonuslinkService.migrations.postgresClientImage }}
  env:
    - name: PGPASSWORD
      valueFrom:
        secretKeyRef:
          name: {{ .Values.bonuslinkService.name }}-secret
          key: postgresPassword
  command:
    - sh
    - -c
    - >-
      until psql -h {{ .Values.bonuslinkService.env.postgresHost }} -p {{ .Values.bonuslinkService.env.postgresPort }}
      -U {{ .Values.bonuslinkService.env.postgresUser | quote }} -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '{{ .Values.bonuslinkService.env.postgresDatabase }}'" | grep -q 1
      || psql -h {{ .Values.bonuslinkService.env.postgresHost }} -p {{ .Values.bonuslinkService.env.postgresPort }}
      -U {{ .Values.bonuslinkService.env.postgresUser | quote }} -d postgres -c "CREATE DATABASE {{ .Values.bonuslinkService.env.postgresDatabase }}";
      do echo "waiting for postgres"; sleep 3; done
- name: migrate
  image: {{ .Values.bonuslinkService.image.name }}:{{ .Values.bonuslinkService.image.tag }}
  imagePullPolicy: Always
  env:
    - name: COM_POSTGRES_PASSWORD
      valueFrom:
        secretKeyRef:
          name: {{ .Values.bonuslinkService.name }}-secret
          key: postgresPassword
  command:
    - sh
    - -c
    - >-
      migrate -path /app/migrations
      -database "postgres://{{ .Values.bonuslinkService.env.postgresUser }}:${COM_POSTGRES_PASSWORD}@{{ .Values.bonuslinkService.env.postgresHost }}:{{ .Values.bonuslinkService.env.postgresPort }}/{{ .Values.bonuslinkService.env.postgresDatabase }}?sslmode={{ .Values.bonuslinkService.env.postgresSSLMode }}"
      up
{{- end }}
{{- end -}}
