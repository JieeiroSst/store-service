{{- define "livestream-service.sidecars" -}}
{{- $node := .Values.livestreamService.node }}
# RTMP ingest sidecar. Lives in the same pod as the app container so
# SRS can forward the raw stream to this node's own ffmpeg process
# over localhost - see internal/adapter/secondary/transcode.
- name: srs
  image: "{{ $node.srs.image }}:{{ $node.srs.tag }}"
  imagePullPolicy: IfNotPresent
  # The image's default CMD loads its own baked-in conf/docker.conf
  # (http_server on 8080 by default), ignoring the srs-conf volume
  # mounted below - which collides with the app container's own
  # port 8080. Must explicitly point at the mounted conf.
  command: ["./objs/srs", "-c", "/usr/local/srs/conf/srs.conf"]
  ports:
  - name: rtmp
    containerPort: {{ $node.srs.rtmpPort }}
  - name: srs-http
    containerPort: {{ $node.srs.httpPort }}
  volumeMounts:
  - name: srs-conf
    mountPath: /usr/local/srs/conf/srs.conf
    subPath: srs.conf
{{- end -}}
