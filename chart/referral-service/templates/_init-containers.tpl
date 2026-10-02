{{- define "referral-service.initContainers" -}}
{{- if .Values.referralService.migrations.enabled }}
- name: create-database
  image: {{ .Values.referralService.migrations.mysqlClientImage }}
  env:
    - name: MYSQL_PWD
      valueFrom:
        secretKeyRef:
          name: {{ .Values.referralService.name }}-secret
          key: mysqlPassword
  command:
    - sh
    - -c
    - >-
      until mysql -h {{ .Values.referralService.env.mysqlHost }} -P {{ .Values.referralService.env.mysqlPort }}
      -u {{ .Values.referralService.env.mysqlUser | quote }} -e "CREATE DATABASE IF NOT EXISTS {{ .Values.referralService.env.mysqlDatabase }}";
      do echo "waiting for mysql"; sleep 3; done
- name: migrate
  image: {{ .Values.referralService.image.name }}:{{ .Values.referralService.image.tag }}
  imagePullPolicy: Always
  env:
    - name: COM_MYSQL_PASSWORD
      valueFrom:
        secretKeyRef:
          name: {{ .Values.referralService.name }}-secret
          key: mysqlPassword
  command:
    - sh
    - -c
    - >-
      migrate -path /app/migrations
      -database "mysql://{{ .Values.referralService.env.mysqlUser }}:${COM_MYSQL_PASSWORD}@tcp({{ .Values.referralService.env.mysqlHost }}:{{ .Values.referralService.env.mysqlPort }})/{{ .Values.referralService.env.mysqlDatabase }}"
      up
{{- end }}
{{- end -}}
