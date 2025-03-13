package synx

import (
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
	"strings"
)

func MDColumnPolicy(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case http.MethodGet:
		action := r.FormValue("action")

		if strings.EqualFold(action, "options") {
			rows, err := asql.Query(tx, "SELECT code, name FROM syn_column_policy ORDER BY order_ ASC")
			if err != nil {
				return nil, err
			}

			res := make([]map[string]string, 0, len(rows))
			for _, row := range rows {
				res = append(res, map[string]string{
					"id":    row["code"],
					"value": row["name"],
				})
			}

			return res, nil
		}

		return asql.Query(tx, "SELECT id, code, name, description, create_at FROM syn_column_policy ORDER BY order_ ASC")
	default:
		operation := r.PostFormValue("operation")

		// 提交数据
		id := r.PostFormValue("id")
		code := r.PostFormValue("code")
		name := r.PostFormValue("name")
		description := r.PostFormValue("description")

		moveId := r.PostFormValue("webix_move_id")
		moveIndex := r.PostFormValue("webix_move_index")
		moveParent := r.PostFormValue("webix_move_parent")

		switch operation {
		case "insert":
			newId, at := asql.GenerateId(), asql.GetDateTime()

			query := "INSERT INTO syn_column_policy(id, code, name, description, order_, create_at) VALUES (?,?,?,?,?,?)"
			args := []interface{}{newId, code, name, description, asql.GenerateOrderId(), at}
			if err := asql.Insert(tx, query, args...); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "newid": newId, "create_at": at}, nil
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
		case "order":
			if err := asql.Order(tx, "syn_column_policy", id, moveId, moveIndex, moveParent); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success"}, nil
		default:
			return nil, fmt.Errorf("unexpect operation %q", operation)
		}
	}
}
