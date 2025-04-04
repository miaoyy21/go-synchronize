
/************************************************ {{ .DstDatabase }}.dbo.{{ .Table }} ************************************************/
{{- if .IsSync }}
    {{- if .Triggers }}
-- 允许 {{ $.DstDatabase }}.dbo.{{ $.Table }} 根据标识插入数据
        {{- range $index, $trigger := .Triggers }}
DISABLE TRIGGER {{ $trigger }} ON {{ $.DstDatabase }}.dbo.{{ $.Table }};
        {{- end }}
    {{ end }}
    {{- if .HasIdentity }}
-- 禁用数据库表 {{ $.DstDatabase }}.dbo.{{ $.Table }} 的所有启用的触发器
SET IDENTITY_INSERT {{ $.DstDatabase }}.dbo.{{ $.Table }} ON;
    {{- end }}

-- 更新 {{ $.DstDatabase }}.dbo.{{ $.Table }} 的数据标识符
UPDATE {{ .DstDatabase }}.dbo.{{ .Table }} SET [_flag_] = '{{ .DstFlag }}' WHERE [_flag_] IS NULL;

-- 将 {{ .SrcDatabase }}.dbo.{{ .Table }} 的数据记录添加到 {{ .DstDatabase }}.dbo.{{ .Table }}
INSERT INTO {{ .DstDatabase }}.dbo.{{ .Table }} (
    {{- range $index, $column := .Columns -}}
        [{{ $column.Name }}]
        {{- if not $column.IsLast}}, {{ end }}
    {{- end -}}
)
SELECT {{ range $index, $column := .Columns }}
        {{- if eq $column.Name "_flag_" }}
    '{{ $.SrcFlag }}'
        {{- else }}
            {{- if eq $column.Policy.Code "Add1000W"  }}
    CASE WHEN ISNULL(T.[{{ $column.Name }}], 0) > 0 THEN T.[{{ $column.Name }}] + 10000000 ELSE T.[{{ $column.Name }}] END /* {{ $column.Policy.Name }} */
            {{- else if eq $column.Policy.Code "SuffixFlag"  }}
    CASE WHEN ISNULL(T.[{{ $column.Name }}], '') <> '' THEN T.[{{ $column.Name }}] + '{{ $.SrcFlag }}' ELSE T.[{{ $column.Name }}] END /* {{ $column.Policy.Name }} */
            {{- else if and (gt (len $column.Policy.ReplaceCode) 1) $column.Policy.IsExactlyMatch  }}
    CASE WHEN X{{ $column.Policy.Index }}.new_value IS NOT NULL THEN X{{ $column.Policy.Index }}.new_value ELSE T.[{{ $column.Name }}] END /* [{{$column.Policy.Index}}].{{ $column.Policy.Name }} */
            {{- else if and (gt (len $column.Policy.ReplaceCode) 1) (not $column.Policy.IsExactlyMatch)  }}
    CASE WHEN ISNULL(T.[{{ $column.Name }}], '') <> '' THEN synchronize.dbo.f_syn_replace('{{ $column.Policy.ReplaceCode }}', T.[{{ $column.Name }}]) ELSE T.[{{ $column.Name }}] END /* {{ $column.Policy.Name }} */
            {{- else }}
    T.[{{ $column.Name -}}]
            {{- end }}
        {{- end -}}
        {{- if not $column.IsLast }},{{ end }}
    {{- end }}
FROM {{ .SrcDatabase }}.dbo.{{ .Table }} T
    {{- range $index, $column := .Columns }}
        {{- if and (gt (len $column.Policy.ReplaceCode) 1) $column.Policy.IsExactlyMatch }}
    LEFT JOIN synchronize.dbo.syn_replace_value X{{ $column.Policy.Index }} ON X{{ $column.Policy.Index }}.code = '{{ $column.Policy.ReplaceCode }}' AND X{{ $column.Policy.Index }}.old_value = T.[{{ $column.Name }}]
        {{- end -}}
    {{ end }}
WHERE NOT EXISTS (SELECT 1 FROM {{ .DstDatabase }}.dbo.{{ .Table }} X WHERE ISNULL(X._flag_, '') = '{{ $.SrcFlag }}');
    {{ if .HasIdentity }}
-- 禁止 {{ $.DstDatabase }}.dbo.{{ $.Table }} 根据标识插入数据
SET IDENTITY_INSERT {{ $.DstDatabase }}.dbo.{{ $.Table }} OFF;
    {{ end }}
    {{- if .Triggers }}
-- 恢复数据库表 {{ $.DstDatabase }}.dbo.{{ $.Table }} 的所有启用的触发器
        {{- range $index, $trigger := .Triggers }}
ENABLE TRIGGER {{ $trigger }} ON {{ $.DstDatabase }}.dbo.{{ $.Table }};
        {{- end }}
    {{ end }}
{{- else }}
-- 警告：数据库表 {{ .SrcDatabase }}.dbo.{{ .Table }} 的迁移同步处于未启用状态，自动忽略
{{ end }}
