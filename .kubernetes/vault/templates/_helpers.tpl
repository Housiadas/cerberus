{{- define "vault.fullname" -}}
cerberus-vault
{{- end -}}

{{- define "vault.labels" -}}
app: {{ include "vault.fullname" . }}
app.kubernetes.io/name: vault
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "vault.selectorLabels" -}}
app: {{ include "vault.fullname" . }}
{{- end -}}
