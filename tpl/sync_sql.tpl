INSERT INTO {{ .DstDatabase }}.dbo.{{ .Table }} (
{{- range $index, $column := .Columns -}}
    {{ $column.Name }}
    {{- if not $column.IsLast}}, {{ end }}
{{- end -}}
)
SELECT {{ range $index, $column := .Columns }}
   {{- if eq $column.Name "_flag_" }}
        '{{ $.SrcFlag }}'
   {{- else }}
        T.{{ $column.Name }}
   {{- end -}}
   {{- if not $column.IsLast}}, {{ end }}
{{- end }}
FROM {{ .SrcDatabase }}.dbo.{{ .Table }} T
WHERE NOT EXISTS (SELECT 1 FROM {{ .DstDatabase }}.dbo.{{ .Table }} X WHERE X._flag_ = '{{ $.SrcFlag }}');
