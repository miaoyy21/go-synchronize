package synx

import (
	"database/sql"
	"go-synchronize/asql"
	"net/http"
)

func MDDatabase(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
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
		dstFlag := r.PostFormValue("dst_flag")
		srcFlag := r.PostFormValue("src_flag")

		switch operation {
		case "insert":
			newId := asql.GenerateId()

			query := "INSERT INTO syn_database(id, dst_db, src_db, dst_flag, src_flag) VALUES (?,?,?,?,?)"
			args := []interface{}{newId, dstDb, srcDb, dstFlag, srcFlag}
			if err := asql.Insert(tx, query, args...); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "newid": newId}, nil
		case "update":
			query := "UPDATE syn_database SET dst_db = ?, src_db = ?, dst_flag = ?, src_flag = ? WHERE id = ?"
			args := []interface{}{dstDb, srcDb, dstFlag, srcFlag, id}
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
