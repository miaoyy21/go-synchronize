
{{- if .IsSync }}
    {{- if or .CreateTable .AddColumn .ModifyColumn }}
        {{- if .CreateTable }}
-- Create Table of {{ .Database }}.dbo.{{ .Table }};
CREATE TABLE {{ .Database }}.dbo.{{ .Table }} (
            {{- range $index, $column := .CreateTable }}
    {{ $column.Name | printf "%-16s" }} {{ $column.Type | printf "%-16s" }} {{ if $column.IsPrimary }}NOT NULL{{ else }}NULL{{ end }},
            {{- end }}
    {{if .Primary.Name}}CONSTRAINT {{ .Table }}_PK PRIMARY KEY ({{ .Primary.Name }}) {{ end }}
);
        {{ end -}}
        {{- if .AddColumn }}
-- Add Columns of {{ .Database }}.dbo.{{ .Table }};
            {{- range $column := .AddColumn }}
ALTER TABLE {{ $.Database }}.dbo.{{ $.Table }} ADD {{ $column.Name }} {{ $column.Type }} NULL;
            {{- end }}
        {{ end -}}
        {{- if .ModifyColumn }}
            {{- range $column := .ModifyColumn }}
-- Modify Column [{{ $column.Name }}]'s data type of {{ $.Database }}.dbo.{{ $.Table }} from {{ $column.TypeOrg }} to {{ $column.Type }};
ALTER TABLE {{ $.Database }}.dbo.{{ $.Table }} ALTER COLUMN {{ $column.Name }} {{ $column.Type }} NULL;
            {{- end }}
        {{ end -}}
    {{- else }}
-- Skip table {{ .Database }}.dbo.{{ .Table }}.
    {{ end -}}
{{- else }}
-- Warning : {{ .Database }}.dbo.{{ .Table }} is NOT Synchronize, and Ignore.
{{ end }}