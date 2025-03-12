

SELECT TT.operation, TT.table_name, TT.column_name, TT.column_type, TT.is_primary, TT.column_type_org
FROM (
	/* [1] 新增数据库表 */
	SELECT 1 AS operation, src.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, NULL AS column_type_org
	FROM syn_table_column src
	WHERE src.database_name = '{{.Src}}'
		AND NOT EXISTS (SELECT 1 FROM syn_table_column dst WHERE dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name)
	UNION ALL 
	/* [2] 新增字段 */
	SELECT 2 AS operation, src.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, NULL AS column_type_org
	FROM syn_table_column src
	WHERE src.database_name = '{{.Src}}'
		AND EXISTS (SELECT 1 FROM syn_table_column dst WHERE dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name)
		AND NOT EXISTS (SELECT 1 FROM syn_table_column dst WHERE dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name AND dst.column_name = src.column_name)
	UNION ALL 
	/* [3] 可安全修改字段 *CHAR *INT *DATE */
	SELECT 3 AS operation, dst.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, dst.column_type AS column_type_org
	FROM syn_table_column src
		INNER JOIN syn_table_column dst ON dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name AND dst.column_name = src.column_name
	WHERE src.database_name = '{{.Src}}' AND dst.column_type <> src.column_type
		AND (
			(
				CHARINDEX('CHAR',dst.column_type,0) > 0 AND CHARINDEX('CHAR',src.column_type,0) > 0
					AND 
				CONVERT(INT,SUBSTRING(dst.column_type,CHARINDEX('(',dst.column_type)+1,CHARINDEX(')',dst.column_type) - CHARINDEX('(',dst.column_type) - 1))
					<
				CONVERT(INT,SUBSTRING(src.column_type,CHARINDEX('(',src.column_type)+1,CHARINDEX(')',src.column_type) - CHARINDEX('(',src.column_type) - 1))
			)
				OR
			(
				dst.column_type = 'TINYINT' 
					AND 
				src.column_type = 'INT'
			)
				OR
			(
				dst.column_type = 'DATE' 
					AND 
				src.column_type = 'DATETIME'
			)
		)
	UNION ALL 
	/* [9] 非安全修改字段  */
	SELECT 9 AS operation, dst.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, dst.column_type AS column_type_org
	FROM syn_table_column src
		INNER JOIN syn_table_column dst ON dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name AND dst.column_name = src.column_name
	WHERE src.database_name = '{{.Src}}' AND dst.column_type <> src.column_type
		AND NOT (
			(
				CHARINDEX('CHAR',dst.column_type,0) > 0 AND CHARINDEX('CHAR',src.column_type,0) > 0
					AND 
				CONVERT(INT,SUBSTRING(dst.column_type,CHARINDEX('(',dst.column_type)+1,CHARINDEX(')',dst.column_type) - CHARINDEX('(',dst.column_type) - 1))
					<
				CONVERT(INT,SUBSTRING(src.column_type,CHARINDEX('(',src.column_type)+1,CHARINDEX(')',src.column_type) - CHARINDEX('(',src.column_type) - 1))
			)
				OR
			(
				dst.column_type = 'TINYINT' 
					AND 
				src.column_type = 'INT'
			)
				OR
			(
				dst.column_type = 'DATE' 
					AND 
				src.column_type = 'DATETIME'
			)
		)
) TT
ORDER BY TT.operation ASC, TT.table_name ASC, TT.column_id ASC;
