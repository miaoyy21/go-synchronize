
SELECT T.name AS table_name,X.name AS trigger_name
FROM {{.Database}}.sys.SYSOBJECTS T
	INNER JOIN {{.Database}}.sys.TRIGGERS X ON X.parent_id  = T.id
WHERE X.is_disabled  <> '1'