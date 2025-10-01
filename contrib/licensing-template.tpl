{
{{- $first := true}}
{{- range . }}
    {{- if $first -}}
        {{- $first = false -}}
        {{- printf "\n" -}}
    {{- else -}}
        {{- printf ",\n" -}}
    {{- end -}}
{{- printf "    " -}}"{{ .Name }}:{{ .Version }}" : "{{ .LicenseName }}"
{{- end }}
}