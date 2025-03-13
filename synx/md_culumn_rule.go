package synx

import (
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
)

func MDColumnRule(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case http.MethodGet:
		return asql.Query(tx, "SELECT * FROM syn_column_rule")
	default:
		operation := r.PostFormValue("operation")

		// 提交数据
		id := r.PostFormValue("id")
		isIgnore := r.PostFormValue("is_ignore")

		switch operation {
		case "update":
			query := "UPDATE syn_column_rule SET is_ignore = ? WHERE id = ?"
			if err := asql.Update(tx, query, isIgnore, id); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "id": id}, nil
		default:
			return nil, fmt.Errorf("unexpect operation %q", operation)
		}
	}
}
