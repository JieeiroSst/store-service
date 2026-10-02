{{- define "car-rental-service.initContainers" -}}
# Wait for postgres and create the database if it does not exist yet.
- name: create-database
  image: {{ .Values.carRentalService.database.postgresClientImage }}
  env:
    - name: PGPASSWORD
      valueFrom:
        secretKeyRef:
          name: {{ .Values.carRentalService.name }}-secret
          key: postgresPassword
  command:
    - sh
    - -c
    - >-
      until psql -h {{ .Values.carRentalService.database.host }} -p {{ .Values.carRentalService.database.port }}
      -U {{ .Values.carRentalService.database.user | quote }} -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '{{ .Values.carRentalService.database.name }}'" | grep -q 1
      || psql -h {{ .Values.carRentalService.database.host }} -p {{ .Values.carRentalService.database.port }}
      -U {{ .Values.carRentalService.database.user | quote }} -d postgres -c "CREATE DATABASE {{ .Values.carRentalService.database.name }}";
      do echo "waiting for postgres"; sleep 3; done
# database.sql ships inside the app image; hand it to the postgres client.
- name: copy-schema
  image: {{ .Values.carRentalService.image.name }}:{{ .Values.carRentalService.image.tag }}
  imagePullPolicy: Always
  command: ["cp", "/app/database.sql", "/schema/database.sql"]
  volumeMounts:
  - name: schema
    mountPath: /schema
# Idempotent (CREATE TABLE IF NOT EXISTS), so safe on every rollout and
# when several replicas start together.
- name: apply-schema
  image: {{ .Values.carRentalService.database.postgresClientImage }}
  env:
    - name: PGPASSWORD
      valueFrom:
        secretKeyRef:
          name: {{ .Values.carRentalService.name }}-secret
          key: postgresPassword
  command:
    - psql
    - -v
    - ON_ERROR_STOP=1
    - -h
    - {{ .Values.carRentalService.database.host | quote }}
    - -p
    - {{ .Values.carRentalService.database.port | quote }}
    - -U
    - {{ .Values.carRentalService.database.user | quote }}
    - -d
    - {{ .Values.carRentalService.database.name | quote }}
    - -f
    - /schema/database.sql
  volumeMounts:
  - name: schema
    mountPath: /schema
{{- end -}}
