
{{- if .IsSync }}
    {{- if or .CreateTable .AddColumn .ModifyColumn }}
        {{- if .CreateTable }}
-- 在数据库 {{ .Database }} 中创建数据库表 {{ .Table }}
CREATE TABLE {{ .Database }}.dbo.{{ .Table }} (
            {{- range $index, $column := .CreateTable }}
    {{ $column.Name | printf "%-16s" }} {{ $column.Type | printf "%-16s" }} {{ if $column.IsPrimary }}NOT NULL{{ else }}NULL{{ end }},
            {{- end }}
    {{ "_flag_" | printf "%-16s" }} {{ "VARCHAR(1)" | printf "%-16s" }} NULL,
    {{if .Primary.Name}}CONSTRAINT {{ .Table }}_PK PRIMARY KEY ({{ .Primary.Name }}) {{ end }}
);
        {{ end -}}
        {{- if .AddColumn }}
-- 在数据库表 {{ .Database }}.dbo.{{ .Table }} 中添加字段
            {{- range $column := .AddColumn }}
ALTER TABLE {{ $.Database }}.dbo.{{ $.Table }} ADD {{ $column.Name }} {{ $column.Type }} NULL;
            {{- end }}
        {{ end -}}
        {{- if .ModifyColumn }}
            {{- range $column := .ModifyColumn }}
-- 修改数据库表 {{ $.Database }}.dbo.{{ $.Table }} [{{ $column.Name }}] 数据类型，由 {{ $column.TypeOrg }} 变成 {{ $column.Type }}
ALTER TABLE {{ $.Database }}.dbo.{{ $.Table }} ALTER COLUMN {{ $column.Name }} {{ $column.Type }} NULL;
            {{- end }}
        {{ end -}}
    {{- else if not $.Flag }}
-- 数据库表 {{ .Database }}.dbo.{{ .Table }} 没有变化，自动忽略
    {{ end -}}
    {{- if $.Flag}}
-- 在 {{ $.Database }}.dbo.{{ $.Table }} 中添加数据库标识符字段
ALTER TABLE {{ $.Database }}.dbo.{{ $.Table }} ADD _flag_ VARCHAR(1) NULL;
UPDATE {{ $.Database }}.dbo.{{ $.Table }} SET _flag_ = '{{ $.Flag }}' WHERE _flag_ IS NULL;
    {{ end -}}
{{- else }}
-- 警告：数据库表 {{ .Database }}.dbo.{{ .Table }} 的迁移同步处于未启用状态，自动忽略
{{ end -}}
