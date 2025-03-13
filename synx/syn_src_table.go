package synx

import (
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
)

func MDSynSrcTable(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
	switch r.Method {
	case http.MethodGet:
		return asql.Query(tx, "SELECT * FROM syn_src_table")
	default:
		operation := r.PostFormValue("operation")

		// 提交数据
		id := r.PostFormValue("id")
		isSync := r.PostFormValue("is_sync")

		switch operation {
		case "update":
			query := "UPDATE syn_src_table SET is_sync = ? WHERE id = ?"
			if err := asql.Update(tx, query, isSync, id); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "id": id}, nil
		default:
			return nil, fmt.Errorf("unexpect operation %q", operation)
		}
	}
}
