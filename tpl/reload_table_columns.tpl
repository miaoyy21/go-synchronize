

        -- 清除历史
        DELETE FROM syn_table_column WHERE database_name = '{{.Database}}';

        -- 更新记录
        INSERT INTO syn_table_column(id, database_name, table_name, column_id, column_name, column_type, column_length, is_primary, is_nullable, is_identity, create_at)
        SELECT NEWID(), '{{.Database}}', TT.table_name, TT.column_id, TT.column_name,
            CASE
                WHEN TT.column_xtype IN ('DECIMAL','NUMERIC') THEN 'NUMERIC'+'('+CONVERT(VARCHAR(10),TT.precision)+','+CONVERT(VARCHAR(10),TT.scale)+')'
                WHEN TT.column_xtype IN ('VARCHAR','CHAR','NVARCHAR','NCHAR','VARBINARY') THEN TT.column_xtype+'('+(CASE WHEN TT.length > 0 THEN CONVERT(VARCHAR(10),TT.length) ELSE 'MAX' END )+')'
            ELSE TT.column_xtype END AS column_type,
            TT.length,TT.is_primary, TT.is_nullable,TT.is_identity,CONVERT(VARCHAR(20),GETDATE(),120)
        FROM (
            SELECT T.name AS table_name,X.column_id,X.name AS column_name,
                CASE X.system_type_id
                    WHEN 34 THEN 'IMAGE'
                    WHEN 35 THEN 'TEXT'
                    WHEN 40 THEN 'DATE'
                    WHEN 48 THEN 'TINYINT'
                    WHEN 56 THEN 'INT'
                    WHEN 61 THEN 'DATETIME'
                    WHEN 62 THEN 'FLOAT'
                    WHEN 99 THEN 'NTEXT'
                    WHEN 106 THEN 'DECIMAL'
                    WHEN 108 THEN 'NUMERIC'
                    WHEN 127 THEN 'BIGINT'
                    WHEN 165 THEN 'VARBINARY'
                    WHEN 167 THEN 'VARCHAR'
                    WHEN 175 THEN 'CHAR'
                    WHEN 231 THEN 'NVARCHAR'
                    WHEN 239 THEN 'NCHAR'
                ELSE CONVERT(VARCHAR(20),X.system_type_id) END AS column_xtype,
                X.max_length AS length,X.precision,X.scale,
                CASE WHEN EXISTS (
                    SELECT 1
                    FROM {{.Database}}.sys.SYSOBJECTS X1
                        INNER JOIN {{.Database}}.sys.SYSINDEXES X2 ON X2.name = X1.name
                        INNER JOIN {{.Database}}.sys.SYSINDEXKEYS X3 ON X3.id = T.id AND X3.indid = X2.indid AND X3.colid = X.column_id
                    WHERE X1.parent_obj = T.id AND X1.xtype = 'PK'
                ) THEN '1' ELSE '0' END AS is_primary,
                X.is_nullable,X.is_identity
            FROM {{.Database}}.sys.SYSOBJECTS T
                LEFT JOIN {{.Database}}.sys.columns X ON X.object_id = T.id
            WHERE T.type = 'U' AND CHARINDEX('_',T.name,0) <> 1
                AND T.name NOT IN ('pbcatcol','pbcatedt','pbcatfmt','pbcattbl','pbcatvld')
        ) TT;
