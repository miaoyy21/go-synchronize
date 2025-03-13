package synx

import (
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
)

func MDSrcTable(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
	switch r.Method {
	case http.MethodGet:
		return asql.Query(tx, "SELECT * FROM syn_src_table WHERE database_name = ? ORDER BY table_name ASC", r.FormValue("database"))
	default:
		operation := r.PostFormValue("operation")

		// 提交数据
		id := r.PostFormValue("id")
		isSync := r.PostFormValue("is_sync")
		description := r.PostFormValue("description")

		switch operation {
		case "update":
			query := "UPDATE syn_src_table SET is_sync = ?, description = ? WHERE id = ?"
			if err := asql.Update(tx, query, isSync, description, id); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "id": id}, nil
		default:
			return nil, fmt.Errorf("unexpect operation %q", operation)
		}
	}
}
