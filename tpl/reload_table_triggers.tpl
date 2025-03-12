

        -- 清除历史
        DELETE FROM syn_table_trigger WHERE database_name = '{{.Database}}';

        -- 更新记录
        INSERT INTO syn_table_trigger(id, database_name, table_name, trigger_name, create_at)
        SELECT NEWID(), '{{.Database}}', T.name AS table_name, X.name AS trigger_name, CONVERT(VARCHAR(20),GETDATE(),120)
        FROM {{.Database}}.sys.SYSOBJECTS T
            INNER JOIN {{.Database}}.sys.TRIGGERS X ON X.parent_id  = T.id
        WHERE X.is_disabled  <> '1';
