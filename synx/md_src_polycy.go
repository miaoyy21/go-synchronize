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
		return asql.Query(tx, "SELECT * FROM syn_src_policy WHERE database_name = ?  AND table_name = ? ORDER BY column_id ASC", database, table)
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
