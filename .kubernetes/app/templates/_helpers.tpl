{{- define "app.fullname" -}}
cerberus-app
{{- end -}}

{{- define "app.labels" -}}
app: {{ include "app.fullname" . }}
app.kubernetes.io/name: cerberus
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "app.selectorLabels" -}}
app: {{ include "app.fullname" . }}
{{- end -}}
