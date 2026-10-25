{{/*
TLS for a single-tenant postgres run through common.app (PCI DSS 4.2).
  - common.postgresTLS.initContainer: issues a self-signed EC P-256 certificate
    into the "postgres-tls" emptyDir on every pod start (clients use
    sslmode=require).
  - common.postgresTLS.hba: pg_hba.conf ConfigMap "<name>-hba" that only
    accepts TLS for TCP clients; the unix socket (probes, initdb) stays trusted.
The postgres container serves them with -c ssl=on, ssl_cert_file and
ssl_key_file under /var/lib/postgresql/tls and hba_file=/etc/postgresql/hba/pg_hba.conf.
Both take (dict "name" <service name> "image" <openssl image>).
*/}}
{{- define "common.postgresTLS.initContainer" -}}
- name: tls-cert
  image: {{ .image | default "alpine/openssl:3.5.6" }}
  command:
    - sh
    - -c
    - |
      set -e
      openssl req -x509 -nodes -days 825 \
        -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
        -subj "/CN={{ .name }}" \
        -addext "subjectAltName=DNS:{{ .name }},DNS:localhost" \
        -keyout /tls/server.key -out /tls/server.crt 2>/dev/null
      chown 999:999 /tls/server.crt /tls/server.key 2>/dev/null || true
      chmod 0644 /tls/server.crt
      chmod 0600 /tls/server.key
  volumeMounts:
    - name: postgres-tls
      mountPath: /tls
{{- end -}}

{{- define "common.postgresTLS.hba" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .name }}-hba
data:
  pg_hba.conf: |
    local     all  all                  trust
    host      all  all  127.0.0.1/32    trust
    host      all  all  ::1/128         trust
    hostssl   all  all  all             scram-sha-256
    hostnossl all  all  all             reject
{{- end -}}
