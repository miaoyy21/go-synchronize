package synx

import (
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
)

func MDColumnPolicy(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case http.MethodGet:
		return asql.Query(tx, "SELECT * FROM syn_column_policy")
	default:
		operation := r.PostFormValue("operation")

		// 提交数据
		id := r.PostFormValue("id")
		code := r.PostFormValue("code")
		name := r.PostFormValue("name")
		description := r.PostFormValue("description")

		switch operation {
		case "insert":
			newId := asql.GenerateId()

			query := "INSERT INTO syn_column_policy(id, code, name, description, create_at) VALUES (?,?,?,?,?)"
			args := []interface{}{newId, code, name, description, asql.GetDateTime()}
			if err := asql.Insert(tx, query, args...); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "newid": newId}, nil
		case "update":
			query := "UPDATE syn_column_policy SET code = ?, name = ?, description = ? WHERE id = ?"
			args := []interface{}{code, name, description, id}
			if err := asql.Update(tx, query, args...); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "id": id}, nil
		case "delete":
			if err := asql.Delete(tx, "DELETE FROM syn_column_policy WHERE id = ?", id); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success"}, nil
		default:
			return nil, fmt.Errorf("unexpect operation %q", operation)
		}
	}
}
