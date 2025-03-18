
{{- if .IsSync }}
    {{- if or .CreateTable .AddColumn .ModifyColumn }}
        {{- if .CreateTable }}
-- Create Table of {{ .Database }}.dbo.{{ .Table }};
CREATE TABLE {{ .Database }}.dbo.{{ .Table }} (
            {{- range $index, $column := .CreateTable }}
    {{ $column.Name | printf "%-16s" }} {{ $column.Type | printf "%-16s" }} {{ if $column.IsPrimary }}NOT NULL{{ else }}NULL{{ end }},
            {{- end }}
    {{ "_flag_" | printf "%-16s" }} {{ "VARCHAR(1)" | printf "%-16s" }} NULL,
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
    {{- else if not $.Flag }}
-- Skip table {{ .Database }}.dbo.{{ .Table }}.
    {{ end -}}
    {{- if $.Flag}}
-- Add Database Flag
ALTER TABLE {{ $.Database }}.dbo.{{ $.Table }} ADD _flag_ VARCHAR(1) NULL;
UPDATE {{ $.Database }}.dbo.{{ $.Table }} SET _flag_ = '{{ $.Flag }}' WHERE _flag_ IS NULL;
    {{ end -}}
{{- else }}
-- Warning : {{ .Database }}.dbo.{{ .Table }} is NOT Synchronize, and Ignore.
{{ end -}}
