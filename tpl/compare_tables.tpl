

        -- 导入字段转换规则
        INSERT INTO syn_column_rule(id, dst_column_type, src_column_type, is_ignore)
        SELECT NEWID(), dst_column_type, src_column_type,
            CASE WHEN (
                ( dst_column_type LIKE 'CHAR%' AND src_column_type LIKE 'CHAR%' AND dst_column_length > src_column_length )
                    OR
                ( dst_column_type LIKE 'NCHAR%' AND src_column_type LIKE 'NCHAR%' AND dst_column_length > src_column_length )
                    OR
                ( dst_column_type LIKE 'VARCHAR%' AND src_column_type LIKE 'VARCHAR%' AND dst_column_length > src_column_length )
                    OR
                ( dst_column_type LIKE 'NVARCHAR%' AND src_column_type LIKE 'NVARCHAR%' AND dst_column_length > src_column_length )
                    OR
                ( dst_column_type = 'INT' AND src_column_type = 'TINYINT' )
                    OR
                ( dst_column_type = 'DATETIME' AND src_column_type = 'DATE')
            ) THEN '1' ELSE '0' END AS is_ignore
        FROM (
            SELECT DISTINCT dst.column_type AS dst_column_type, dst.column_length AS dst_column_length,
                src.column_type AS src_column_type, src.column_length AS src_column_length
            FROM syn_table_column src
                INNER JOIN syn_table_column dst ON dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name AND dst.column_name = src.column_name
            WHERE src.database_name = '{{.Src}}' AND dst.column_type <> src.column_type
                AND NOT EXISTS (SELECT 1 FROM syn_column_rule rule WHERE rule.dst_column_type = dst.column_type AND rule.src_column_type = src.column_type)
        );

        -- 差异对比
        SELECT TT.operation, TT.table_name, TT.column_name, TT.column_type, TT.is_primary, TT.column_type_org
        FROM (
            /* [1] 新增数据库表 */
            SELECT '1' AS operation, src.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, NULL AS column_type_org
            FROM syn_table_column src
            WHERE src.database_name = '{{.Src}}'
                AND NOT EXISTS (SELECT 1 FROM syn_table_column dst WHERE dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name)
            UNION ALL
            /* [2] 新增字段 */
            SELECT '2' AS operation, src.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, NULL AS column_type_org
            FROM syn_table_column src
            WHERE src.database_name = '{{.Src}}'
                AND EXISTS (SELECT 1 FROM syn_table_column dst WHERE dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name)
                AND NOT EXISTS (SELECT 1 FROM syn_table_column dst WHERE dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name AND dst.column_name = src.column_name)
            UNION ALL
            /* [3] 类型修改 */
            SELECT '3' AS operation, dst.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, dst.column_type AS column_type_org
            FROM syn_table_column src
                INNER JOIN syn_table_column dst ON dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name AND dst.column_name = src.column_name
            WHERE src.database_name = '{{.Src}}' AND dst.column_type <> src.column_type
                AND NOT EXISTS (SELECT 1 FROM syn_column_rule rule WHERE rule.dst_column_type = dst.column_type AND rule.src_column_type = src.column_type AND rule.is_ignore = '1')
        ) TT
        ORDER BY TT.operation ASC, TT.table_name ASC, TT.column_id ASC
