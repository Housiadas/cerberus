{{- define "tempo.fullname" -}}
cerberus-tempo
{{- end -}}

{{- define "tempo.labels" -}}
app: {{ include "tempo.fullname" . }}
app.kubernetes.io/name: tempo
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "tempo.selectorLabels" -}}
app: {{ include "tempo.fullname" . }}
{{- end -}}
