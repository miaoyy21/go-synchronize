

        /**************************************************************** 查询数据库表结构差异 ****************************************************************/
        -- 删除原始数据库表结构差异
        DELETE FROM syn_src_difference WHERE database_name = '{{.Src}}';

        -- 导入原始数据库表结构差异
        INSERT INTO syn_src_difference(id, difference_type, database_name, table_name, column_id, column_name, column_type, is_primary, column_type_org, create_at)
        SELECT NEWID(), TT.difference_type, '{{.Src}}', TT.table_name, TT.column_id, TT.column_name, TT.column_type, TT.is_primary, TT.column_type_org, CONVERT(VARCHAR(20),GETDATE(),120)
        FROM (
            /* [1] 新增数据库表 */
            SELECT 'Create Table' AS difference_type, src.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, NULL AS column_type_org
            FROM syn_table_column src
            WHERE src.database_name = '{{.Src}}'
                AND NOT EXISTS (SELECT 1 FROM syn_table_column dst WHERE dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name)
            UNION ALL
            /* [2] 新增字段 */
            SELECT 'Add Column' AS difference_type, src.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, NULL AS column_type_org
            FROM syn_table_column src
            WHERE src.database_name = '{{.Src}}'
                AND EXISTS (SELECT 1 FROM syn_table_column dst WHERE dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name)
                AND NOT EXISTS (SELECT 1 FROM syn_table_column dst WHERE dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name AND dst.column_name = src.column_name)
            UNION ALL
            /* [3] 类型修改 */
            SELECT 'Modify Column' AS difference_type, dst.table_name, src.column_id, src.column_name, src.column_type, src.is_primary, dst.column_type AS column_type_org
            FROM syn_table_column src
                INNER JOIN syn_table_column dst ON dst.database_name = '{{.Dst}}' AND dst.table_name = src.table_name AND dst.column_name = src.column_name
            WHERE src.database_name = '{{.Src}}' AND dst.column_type <> src.column_type
                AND NOT EXISTS (SELECT 1 FROM syn_column_rule ruler WHERE ruler.dst_column_type = dst.column_type AND ruler.src_column_type = src.column_type AND ruler.is_ignore = '1')
        ) TT
        ORDER BY TT.difference_type ASC, TT.table_name ASC, TT.column_id ASC;
