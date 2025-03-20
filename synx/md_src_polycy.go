package synx

import (
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
)

func MdSrcPolicy(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
	switch r.Method {
	case http.MethodGet:
		database, table := r.FormValue("database_name"), r.FormValue("table_name")
		return asql.Query(tx, `
			SELECT T.id, T.column_name, T.column_type, T.column_policy, X.is_primary, X.is_identity, X.is_nullable, T.create_at 
			FROM syn_src_policy T 
				LEFT JOIN syn_table_column X ON X.database_name = T.database_name AND X.table_name = T.table_name AND X.column_name = T.column_name 
			WHERE T.database_name = ?  AND T.table_name = ? ORDER BY T.column_id ASC
		`, database, table)
	default:
		operation := r.PostFormValue("operation")

		// 提交数据
		id := r.PostFormValue("id")
		columnPolicy := r.PostFormValue("column_policy")

		switch operation {
		case "update":
			query := "UPDATE syn_src_policy SET column_policy = ? WHERE id = ?"
			if err := asql.Update(tx, query, columnPolicy, id); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "id": id}, nil
		default:
			return nil, fmt.Errorf("unexpect operation %q", operation)
		}
	}
}
