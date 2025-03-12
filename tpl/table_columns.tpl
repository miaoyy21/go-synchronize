SELECT TT.TABLE_NAME,TT.COLUMN_NAME,
	CASE
		WHEN TT.COLUMN_XTYPE IN ('DECIMAL','NUMERIC') THEN TT.COLUMN_XTYPE+'('+CONVERT(VARCHAR(10),TT.xprec)+','+CONVERT(VARCHAR(10),TT.xscale)+')'
		WHEN TT.COLUMN_XTYPE IN ('VARCHAR','CHAR','NVARCHAR','NCHAR') THEN TT.COLUMN_XTYPE+'('+CONVERT(VARCHAR(10),TT.length)+')'
	ELSE TT.COLUMN_XTYPE END AS COLUMN_TYPE,
	TT.IS_PRIMARY
FROM (
	SELECT T.id AS TABLE_ID,T.name AS TABLE_NAME,
		X.colid AS COLUMN_ID,X.name AS COLUMN_NAME,
		CASE X.xtype
			WHEN 34 THEN 'IMAGE'
			WHEN 35 THEN 'TEXT'
			WHEN 40 THEN 'DATE'
			WHEN 48 THEN 'TINYINT'
			WHEN 56 THEN 'INT'
			WHEN 61 THEN 'DATETIME'
			WHEN 62 THEN 'FLOAT'
			WHEN 106 THEN 'DECIMAL'
			WHEN 108 THEN 'NUMERIC'
			WHEN 167 THEN 'VARCHAR'
			WHEN 175 THEN 'CHAR'
			WHEN 231 THEN 'NVARCHAR'
			WHEN 239 THEN 'NCHAR'
		ELSE CONVERT(VARCHAR(20),X.xtype) END AS COLUMN_XTYPE,
		X.length,X.xprec,X.xscale,
		CASE WHEN EXISTS (
			SELECT 1
			FROM {{.Database}}.sys.SYSOBJECTS X1
				INNER JOIN {{.Database}}.sys.SYSINDEXES X2 ON X2.name = X1.name
				INNER JOIN {{.Database}}.sys.SYSINDEXKEYS X3 ON X3.id = T.id AND X3.indid = X2.indid AND X3.colid = X.colid
			WHERE X1.parent_obj = T.id AND X1.xtype = 'PK'
		) THEN '1' ELSE '0' END AS IS_PRIMARY
	FROM {{.Database}}.sys.SYSOBJECTS T
		LEFT JOIN {{.Database}}.sys.SYSCOLUMNS X ON X.id = T.id
	WHERE T.type = 'U' AND CHARINDEX('_',T.name,0) <> 1
		AND T.name NOT IN ('pbcatcol','pbcatedt','pbcatfmt','pbcattbl','pbcatvld')
) TT

