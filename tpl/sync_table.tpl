
{{- if .IsSync }}
    {{- if or .CreateTable .AddColumn .ModifyColumn }}
        {{- if .CreateTable }}
-- 在数据库 {{ .Database }} 中创建数据库表 {{ .Table }}
IF OBJECT_ID('{{ .Table }}', 'U') IS NULL
    CREATE TABLE {{ .Database }}.dbo.{{ .Table }} (
                {{- range $index, $column := .CreateTable }}
        {{ $column.Name | printf "[%s]" | printf "%-16s" }} {{ $column.Type | printf "%-16s" }} {{if $column.IsIdentity}} IDENTITY(1,1) {{ end }} {{ if $column.IsNullable }}NULL{{ else }}NOT NULL{{ end }},
                {{- end }}
        {{ "_flag_" | printf "[%s]" | printf "%-16s" }} {{ "VARCHAR(1)" | printf "%-16s" }}  NULL,
        {{if .Primary.Name}}CONSTRAINT {{ .Table }}_PK PRIMARY KEY ([{{ .Primary.Name }}]) {{ end }}
    );
            {{ end -}}
        {{- if .AddColumn }}
-- 在数据库表 {{ .Database }}.dbo.{{ .Table }} 中添加字段
            {{- range $column := .AddColumn }}
IF NOT EXISTS (SELECT 1 FROM {{ $.Database }}.sys.columns WHERE object_id = OBJECT_ID('{{ $.Table }}') AND name = '{{ $column.Name }}')
    ALTER TABLE {{ $.Database }}.dbo.{{ $.Table }} ADD [{{ $column.Name }}] {{ $column.Type }} {{ if $column.IsNullable }}NULL{{ else }}NOT NULL{{ end }};
            {{- end }}
        {{ end -}}
        {{- if .ModifyColumn }}
            {{- range $column := .ModifyColumn }}
-- 修改数据库表 {{ $.Database }}.dbo.{{ $.Table }} [{{ $column.Name }}] 数据类型，由 {{ $column.TypeOrg }} 变成 {{ $column.Type }}
ALTER TABLE {{ $.Database }}.dbo.{{ $.Table }} ALTER COLUMN [{{ $column.Name }}] {{ $column.Type }} {{ if $column.IsNullable }}NULL{{ else }}NOT NULL{{ end }};
            {{- end }}
        {{ end -}}
    {{- else if not $.Flag }}
-- 数据库表 {{ .Database }}.dbo.{{ .Table }} 没有变化，自动忽略
    {{ end -}}
    {{- if $.Flag}}
-- 在 {{ $.Database }}.dbo.{{ $.Table }} 中添加数据库标识符字段
IF NOT EXISTS (SELECT 1 FROM {{ $.Database }}.sys.columns WHERE object_id = OBJECT_ID('{{ $.Table }}') AND name = '_flag_')
ALTER TABLE {{ $.Database }}.dbo.{{ $.Table }} ADD [_flag_] VARCHAR(1) NULL;
    {{ end -}}
{{- else }}
-- 警告：数据库表 {{ .Database }}.dbo.{{ .Table }} 的迁移同步处于未启用状态，自动忽略
{{ end -}}
