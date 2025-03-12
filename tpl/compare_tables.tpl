/* 新增数据库表 */
SELECT 1 AS operation, src.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, NULL AS column_type_org
FROM syn_table_column src
WHERE src.database_name = 'jz23_scxt'
	AND NOT EXISTS (SELECT 1 FROM syn_table_column dst WHERE dst.database_name = 'jz24_scxt' AND dst.table_name = src.table_name)
UNION ALL
/* 新增字段 */
SELECT 2 AS operation, src.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, NULL AS column_type_org
FROM syn_table_column src
WHERE src.database_name = 'jz23_scxt'
	AND EXISTS (SELECT 1 FROM syn_table_column dst WHERE dst.database_name = 'jz24_scxt' AND dst.table_name = src.table_name)
	AND NOT EXISTS (SELECT 1 FROM syn_table_column dst WHERE dst.database_name = 'jz24_scxt' AND dst.table_name = src.table_name AND dst.column_name = src.column_name)
UNION ALL
/* 可安全修改字段 */
SELECT 3 AS operation, dst.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, dst.column_type AS column_type_org
FROM syn_table_column src
	INNER JOIN syn_table_column dst ON dst.database_name = 'jz24_scxt' AND dst.table_name = src.table_name AND dst.column_name = src.column_name
WHERE src.database_name = 'jz23_scxt' AND dst.column_type <> src.column_type
	AND CHARINDEX('CHAR',dst.column_type,0) > 0 AND CHARINDEX('CHAR',src.column_type,0) > 0