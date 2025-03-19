

        /**************************************************************** 初始化 ****************************************************************/
        -- syn_replace_code
        INSERT INTO syn_replace_code(id, code, name, description, order_, create_at)
        SELECT TT.id, TT.code, TT.name, TT.description, TT.order_, TT.create_at
        FROM (
            SELECT '0' AS id, '-' AS code, '-' AS name, '数据库完整性约束，不启用替换模式' AS description, 10000 AS order_, CONVERT(VARCHAR(20),GETDATE(),120) AS create_at
            UNION ALL
            SELECT '1', 'RYDM', '人员代码', '人员编号替换', 20000, CONVERT(VARCHAR(20),GETDATE(),120)
        ) TT
        WHERE NOT EXISTS (SELECT 1 FROM syn_replace_code XX WHERE XX.code = TT.code);

        -- syn_column_policy
        INSERT INTO syn_column_policy(id, code, name, description, create_at, order_, is_exactly_match, replace_code)
        SELECT id, code, name, description, create_at, order_, is_exactly_match, replace_code
        FROM (
            SELECT '0', 'None', '-', '系统默认添加的字段更新策略，表示该字段没设置更新策略', CONVERT(VARCHAR(20),GETDATE(),120), 10000, '0', '-'
            UNION ALL
            SELECT '1' AS id, 'Add1000W' AS code, '数值增加1千万' AS name, '处理自增数值ID或关联的数值ID，忽略零值【0】。示例：123 => 10000123' AS description, CONVERT(VARCHAR(20),GETDATE(),120) AS create_at, 20000 AS order_, '0' AS is_exactly_match, '-' AS replace_code
            UNION ALL
            SELECT '2', 'SuffixFlag', '尾部添加标识符', '处理单据号，忽略零值【''''】。示例：HTR250001 => HTR250001Y', CONVERT(VARCHAR(20),GETDATE(),120), 30000, '0', '-'
            UNION ALL
            SELECT '3', 'ReplaceRybh', '替换人员编号', '替换人员编号，忽略零值【''''】。示例：WZB250001 => 0120250001', CONVERT(VARCHAR(20),GETDATE(),120), 40000, '1', 'RYDM'
            UNION ALL
            SELECT '4', 'ReplaceRybhEx', '替换人员编号【多】', '替换多个人员编号，忽略零值【''''】。示例：WZB250001,WZB250003 => 0120250001,0120250003', CONVERT(VARCHAR(20),GETDATE(),120), 50000, '0', 'RYDM'
        ) TT
        WHERE NOT EXISTS (SELECT 1 FROM syn_column_policy XX WHERE XX.code = TT.code)

        /**************************************************************** 原始数据库表 ****************************************************************/
        -- 删除没有的原始数据库表
        DELETE syn
        FROM syn_src_table syn
        WHERE NOT EXISTS (SELECT 1 FROM syn_table_column src WHERE src.database_name = syn.database_name AND src.table_name = syn.table_name);

        -- 导入原始数据库表
        INSERT INTO syn_src_table(id, database_name, table_name, is_sync, rows, create_at)
        SELECT NEWID(), TT.database_name, TT.table_name, '1' AS is_sync, XX.rows, CONVERT(VARCHAR(20),GETDATE(),120)
        FROM (
            SELECT DISTINCT src.database_name, src.table_name
            FROM syn_table_column src
            WHERE src.database_name = '{{.Src}}'
                AND NOT EXISTS (SELECT 1 FROM syn_src_table syn WHERE syn.database_name = src.database_name AND syn.table_name = src.table_name)
        ) TT
            LEFT JOIN syn_table XX ON XX.database_name = TT.database_name AND XX.table_name = TT.table_name;

        /**************************************************************** 原始数据库表的策略 ****************************************************************/
        -- 删除没有的原始数据库表的策略
        DELETE syn
        FROM syn_src_policy syn
        WHERE NOT EXISTS (SELECT 1 FROM syn_table_column src WHERE src.database_name = syn.database_name AND src.table_name = syn.table_name AND src.column_name = syn.column_name);

        -- 导入原始数据库表的策略
        INSERT INTO syn_src_policy(id, database_name, table_name, column_id, column_name, column_type, is_primary, column_policy, create_at)
        SELECT NEWID(), src.database_name, src.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, 'None', CONVERT(VARCHAR(20),GETDATE(),120)
        FROM syn_table_column src
        WHERE src.database_name = '{{.Src}}'
            AND NOT EXISTS (SELECT 1 FROM syn_src_policy syn WHERE syn.database_name = src.database_name AND syn.table_name = src.table_name AND syn.column_name = src.column_name);

        /**************************************************************** 字段转换规则 ****************************************************************/
        -- 导入字段转换规则
        INSERT INTO syn_column_rule(id, dst_column_type, src_column_type, is_ignore, create_at)
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
                ( dst_column_type = 'BIGINT' AND src_column_type LIKE '%INT%' )
                    OR
                ( dst_column_type = 'DATETIME' AND src_column_type = 'DATE')
            ) THEN '1' ELSE '0' END AS is_ignore, CONVERT(VARCHAR(20),GETDATE(),120)
        FROM (
            SELECT DISTINCT dst.column_type AS dst_column_type, dst.column_length AS dst_column_length,
                src.column_type AS src_column_type, src.column_length AS src_column_length
            FROM syn_table_column src
                INNER JOIN syn_table_column dst ON dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name AND dst.column_name = src.column_name
            WHERE src.database_name = '{{.Src}}' AND dst.column_type <> src.column_type
                AND NOT EXISTS (SELECT 1 FROM syn_column_rule ruler WHERE ruler.dst_column_type = dst.column_type AND ruler.src_column_type = src.column_type)
        ) TT;
