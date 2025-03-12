package synx

import (
	"database/sql"
	"go-synchronize/asql"
	"net/http"
)

func MDDataBase(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
	switch r.Method {
	case http.MethodGet:
		// 查询请求
		return asql.Query(tx, "SELECT * FROM syn_database ")
	default:
		// 提交请求
		operation := r.PostFormValue("operation")

		id := r.PostFormValue("id")
		dstDb := r.PostFormValue("dst_db")
		srcDb := r.PostFormValue("src_db")

		switch operation {
		case "insert":
			newId := asql.GenerateId()

			query := "INSERT INTO syn_database(id, dst_db, src_db) VALUES (?,?,?)"
			args := []interface{}{newId, dstDb, srcDb}
			if err := asql.Insert(tx, query, args...); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "newid": newId}, nil
		case "update":
			query := "UPDATE syn_database SET dst_db = ?, src_db = ? WHERE id = ?"
			args := []interface{}{dstDb, srcDb, id}
			if err := asql.Update(tx, query, args...); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "id": id}, nil
		case "delete":
			if err := asql.Delete(tx, "DELETE FROM syn_database WHERE id = ?", id); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success"}, nil
		default:
			return res, nil
		}
	}
}
