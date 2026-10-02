{{/*
Every template in this library takes the same context:

  (dict "root" $ "svc" .Values.<serviceKey>)

"root" is the calling chart's top-level context, "svc" is the service's values
block (see chart/vending-machine-service/values.yaml for the full contract).
*/}}

{{- define "common.name" -}}
{{- .svc.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "common.environment" -}}
{{- $global := .root.Values.global | default dict -}}
{{- $global.environment | default "dev" -}}
{{- end -}}

{{/*
Selector labels. Kept as the bare `app` label the existing charts already use
(`selectorApp` overrides it, most charts use "<name>-deployment"): a
Deployment's selector is immutable, so changing it would block upgrades of
releases that are already running.
*/}}
{{- define "common.selectorLabels" -}}
app: {{ .svc.selectorApp | default (include "common.name" .) }}
{{- end -}}

{{- define "common.labels" -}}
{{ include "common.selectorLabels" . }}
app.kubernetes.io/name: {{ include "common.name" . }}
app.kubernetes.io/instance: {{ .root.Release.Name }}
app.kubernetes.io/version: {{ .svc.image.tag | default .root.Chart.AppVersion | toString | trunc 63 | quote }}
app.kubernetes.io/part-of: store-service
app.kubernetes.io/managed-by: {{ .root.Release.Service }}
helm.sh/chart: {{ printf "%s-%s" .root.Chart.Name .root.Chart.Version | replace "+" "_" }}
environment: {{ include "common.environment" . }}
{{- end -}}

{{/*
Labels the platform chart's shared NetworkPolicies key on (chart/platform).
*/}}
{{- define "common.networkLabels" -}}
{{- $net := .svc.networkLabels | default dict -}}
{{- if $net.consulClient }}
network.store.io/consul-client: "true"
{{- end }}
{{- $exposed := (.svc.ingress | default dict).enabled -}}
{{- range .svc.extraIngresses }}{{ if .enabled }}{{ $exposed = true }}{{ end }}{{ end }}
{{- if $exposed }}
network.store.io/ingress: "true"
{{- end }}
{{- if (.svc.metrics | default dict).enabled }}
network.store.io/metrics: "true"
{{- end }}
{{- if $net.internet }}
network.store.io/internet: "true"
{{- end }}
{{- if $net.isolated }}
network.store.io/isolated: "true"
{{- end }}
{{- end -}}

{{- define "common.serviceAccountName" -}}
{{- $sa := .svc.serviceAccount | default dict -}}
{{- if $sa.create -}}
{{- $sa.name | default (include "common.name" .) -}}
{{- else -}}
{{- $sa.name | default "default" -}}
{{- end -}}
{{- end -}}

{{/*
Image reference. A digest, when set, wins over the tag so production can pin
the exact image that was scanned and promoted.
*/}}
{{- define "common.image" -}}
{{- $global := .root.Values.global | default dict -}}
{{- $repo := .svc.image.name -}}
{{- with $global.imageRegistry -}}
{{- $repo = printf "%s/%s" . $repo -}}
{{- end -}}
{{- if .svc.image.digest -}}
{{- printf "%s@%s" $repo .svc.image.digest -}}
{{- else -}}
{{- printf "%s:%s" $repo (.svc.image.tag | toString) -}}
{{- end -}}
{{- end -}}

{{/*
Workload object name. Defaults to "<name>-deployment"; charts whose existing
objects are named differently set `deploymentName` so upgrades replace the
object in place instead of creating a second one next to it.
*/}}
{{- define "common.workloadName" -}}
{{- .svc.deploymentName | default (printf "%s-deployment" (include "common.name" .)) -}}
{{- end -}}

{{- define "common.workloadKind" -}}
{{- if .svc.statefulSet }}StatefulSet{{ else }}Deployment{{ end -}}
{{- end -}}

{{- define "common.serviceName" -}}
{{- (.svc.service | default dict).name | default (printf "%s-svc" (include "common.name" .)) -}}
{{- end -}}

{{- define "common.servicePort" -}}
{{- (.svc.service | default dict).port | default 80 -}}
{{- end -}}

{{/*
Renders every resource of a standard HTTP service. Charts that need nothing
custom only have to include this one template.
*/}}
{{- define "common.app" -}}
{{ include "common.serviceaccount" . }}
{{ include "common.secret" . }}
{{ include "common.consulSeed" . }}
{{ include "common.deployment" . }}
{{ include "common.service" . }}
{{ include "common.ingress" . }}
{{ include "common.hpa" . }}
{{ include "common.pdb" . }}
{{ include "common.networkpolicy" . }}
{{ include "common.servicemonitor" . }}
{{ include "common.externalsecret" . }}
{{- end -}}

{{/*
A background workload (queue worker, consumer): same pod hardening, scaling
and disruption budget as common.app, but no Service, Ingress or metrics
scraping. Leave image.port unset when the process listens on nothing.
*/}}
{{- define "common.workload" -}}
{{ include "common.serviceaccount" . }}
{{ include "common.secret" . }}
{{ include "common.consulSeed" . }}
{{ include "common.deployment" . }}
{{ include "common.hpa" . }}
{{ include "common.pdb" . }}
{{ include "common.networkpolicy" . }}
{{ include "common.externalsecret" . }}
{{- end -}}
