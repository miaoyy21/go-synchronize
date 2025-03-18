
/************************************************ {{ .DstDatabase }}.dbo.{{ .Table }} ************************************************/
{{- if .Triggers }}
    {{- range $index, $trigger := .Triggers }}
-- 禁用数据库表 {{ $.DstDatabase }}.dbo.{{ $.Table }} 的触发器 {{ $trigger }}
DISABLE TRIGGER {{ $trigger }} ON {{ $.DstDatabase }}.dbo.{{ $.Table }};
    {{ end }}
{{- end }}
-- 将 {{ .SrcDatabase }}.dbo.{{ .Table }} 的数据记录添加到 {{ .DstDatabase }}.dbo.{{ .Table }}
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
        {{- if eq $column.Policy.Code "Add1000W"  }}
        CASE WHEN ISNULL(CONVERT({{ $column.Type }}, T.{{ $column.Name }}), 0) > 0 THEN CONVERT({{ $column.Type }}, T.{{ $column.Name }}) + 10000000 ELSE T.{{ $column.Name }} END /* {{ $column.Policy.Name }} */
        {{- else if eq $column.Policy.Code "SuffixFlag"  }}
        CASE WHEN ISNULL(CONVERT({{ $column.Type }}, T.{{ $column.Name }}), '') <> '' THEN CONVERT({{ $column.Type }}, T.{{ $column.Name }}) + '{{ $.SrcFlag }}' ELSE T.{{ $column.Name }} END /* {{ $column.Policy.Name }} */
        {{- else }}
        T.{{ $column.Name -}}
        {{ end }}
   {{- end -}}
   {{- if not $column.IsLast }},{{ end }}
{{- end }}
FROM {{ .SrcDatabase }}.dbo.{{ .Table }} T
WHERE NOT EXISTS (SELECT 1 FROM {{ .DstDatabase }}.dbo.{{ .Table }} X WHERE X._flag_ = '{{ $.SrcFlag }}');
{{ if .Triggers }}
    {{- range $index, $trigger := .Triggers }}
-- 启用数据库表 {{ $.DstDatabase }}.dbo.{{ $.Table }} 的触发器 {{ $trigger }}
ENABLE TRIGGER {{ $trigger }} ON {{ $.DstDatabase }}.dbo.{{ $.Table }};
    {{ end }}
{{- end -}}
