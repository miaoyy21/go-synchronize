
/************************************************ {{ .DstDatabase }}.dbo.{{ .Table }} ************************************************/
{{- if .IsSync }}
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
    CASE WHEN ISNULL(T.{{ $column.Name }}, 0) > 0 THEN T.{{ $column.Name }} + 10000000 ELSE T.{{ $column.Name }} END /* {{ $column.Policy.Name }} */
            {{- else if eq $column.Policy.Code "SuffixFlag"  }}
    CASE WHEN ISNULL(T.{{ $column.Name }}, '') <> '' THEN T.{{ $column.Name }} + '{{ $.SrcFlag }}' ELSE T.{{ $column.Name }} END /* {{ $column.Policy.Name }} */
            {{- else if and $column.Policy.IsUsingReplace $column.Policy.IsExactlyMatch  }}
    CASE WHEN X{{ $column.Policy.Index }}.new_value IS NOT NULL THEN X{{ $column.Policy.Index }}.new_value ELSE T.{{ $column.Name }} END /* {{ $column.Policy.Name }} */
            {{- else }}
    T.{{ $column.Name -}}
            {{ end }}
        {{- end -}}
        {{- if not $column.IsLast }},{{ end }}
    {{- end }}
FROM {{ .SrcDatabase }}.dbo.{{ .Table }} T
    {{- range $index, $column := .Columns }}
        {{- if and $column.Policy.IsUsingReplace $column.Policy.IsExactlyMatch }}
    LEFT JOIN syn_replace X{{ $column.Policy.Index }} ON X{{ $column.Policy.Index }}.policy_code = '{{ $column.Policy.Code }}' AND X{{ $column.Policy.Index }}.old_value = T.{{ $column.Name }}
        {{- end -}}
    {{ end }}
WHERE NOT EXISTS (SELECT 1 FROM {{ .DstDatabase }}.dbo.{{ .Table }} X WHERE ISNULL(X._flag_, '') = '{{ $.SrcFlag }}');
    {{ if .Triggers }}
        {{- range $index, $trigger := .Triggers }}
-- 启用数据库表 {{ $.DstDatabase }}.dbo.{{ $.Table }} 的触发器 {{ $trigger }}
ENABLE TRIGGER {{ $trigger }} ON {{ $.DstDatabase }}.dbo.{{ $.Table }};
        {{ end }}
    {{- end -}}
{{- else }}
-- 警告：数据库表 {{ .SrcDatabase }}.dbo.{{ .Table }} 的迁移同步处于未启用状态，自动忽略
{{ end -}}
