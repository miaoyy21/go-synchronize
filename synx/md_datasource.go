package synx

import (
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
	"strings"
)

func MDDatasource(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case http.MethodGet:
		action := r.FormValue("action")
		if strings.EqualFold(action, "options") {
			return asql.Query(tx, "SELECT code AS id, name AS value FROM syn_datasource ORDER BY order_ ASC")
		}

		return asql.Query(tx, "SELECT * FROM syn_datasource ORDER BY order_ ASC")
	default:
		operation := r.PostFormValue("operation")

		// 提交数据
		id := r.PostFormValue("id")
		code := r.PostFormValue("code")
		name := r.PostFormValue("name")
		driver := r.PostFormValue("driver")
		datasource := r.PostFormValue("datasource")

		moveId := r.PostFormValue("webix_move_id")
		moveIndex := r.PostFormValue("webix_move_index")
		moveParent := r.PostFormValue("webix_move_parent")

		switch operation {
		case "insert":
			newId, at := asql.GenerateId(), asql.GetDateTime()

			query := "INSERT INTO syn_datasource(id, code, name, driver, datasource, order_, create_at) VALUES (?,?,?,?,?,?,?)"
			args := []interface{}{newId, code, name, driver, datasource, asql.GenerateOrderId(), at}
			if err := asql.Insert(tx, query, args...); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "newid": newId, "create_at": at}, nil
		case "update":
			query := "UPDATE syn_datasource SET code = ?, name = ?, driver = ?, datasource = ? WHERE id = ?"
			args := []interface{}{code, name, driver, datasource, id}
			if err := asql.Update(tx, query, args...); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "id": id}, nil
		case "delete":
			if err := asql.Delete(tx, "DELETE FROM syn_datasource WHERE id = ?", id); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success"}, nil
		case "order":
			if err := asql.Order(tx, "syn_datasource", id, moveId, moveIndex, moveParent); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success"}, nil
		default:
			return nil, fmt.Errorf("unexpect operation %q", operation)
		}
	}
}
