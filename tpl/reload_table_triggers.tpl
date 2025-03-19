

        -- 清除历史
        DELETE FROM syn_table WHERE database_name = '{{.Database}}';

        -- 更新表的记录数
        INSERT INTO syn_table(id, database_name, table_name, rows, create_at)
        SELECT NEWID(), '{{.Database}}', TT.name, TT.rows, CONVERT(VARCHAR(20),GETDATE(),120)
        FROM (
            SELECT T.name, SUM(X.rows) AS rows
            FROM {{.Database}}.sys.SYSOBJECTS T
                LEFT JOIN {{.Database}}.sys.PARTITIONS X ON X.object_id = T.id
            WHERE T.type = 'U' AND CHARINDEX('_',T.name,0) <> 1
                  AND T.name NOT IN ('pbcatcol','pbcatedt','pbcatfmt','pbcattbl','pbcatvld') AND X.index_id IN (0, 1)
            GROUP BY T.name
        ) TT;

        -- 清除历史
        DELETE FROM syn_table_trigger WHERE database_name = '{{.Database}}';

        -- 更新记录
        INSERT INTO syn_table_trigger(id, database_name, table_name, trigger_name, create_at)
        SELECT NEWID(), '{{.Database}}', T.name AS table_name, X.name AS trigger_name, CONVERT(VARCHAR(20),GETDATE(),120)
        FROM {{.Database}}.sys.SYSOBJECTS T
            INNER JOIN {{.Database}}.sys.TRIGGERS X ON X.parent_id  = T.id
        WHERE X.is_disabled  <> '1';
