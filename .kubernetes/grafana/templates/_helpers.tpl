{{- define "grafana.fullname" -}}
cerberus-grafana
{{- end -}}

{{- define "grafana.labels" -}}
app: {{ include "grafana.fullname" . }}
app.kubernetes.io/name: grafana
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "grafana.selectorLabels" -}}
app: {{ include "grafana.fullname" . }}
{{- end -}}
