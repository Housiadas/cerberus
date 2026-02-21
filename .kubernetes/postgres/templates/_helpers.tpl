{{- define "postgres.fullname" -}}
cerberus-postgres
{{- end -}}

{{- define "postgres.labels" -}}
app: {{ include "postgres.fullname" . }}
app.kubernetes.io/name: postgres
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "postgres.selectorLabels" -}}
app: {{ include "postgres.fullname" . }}
{{- end -}}
