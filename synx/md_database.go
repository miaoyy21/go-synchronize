package synx

import (
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
	"strings"
)

func MDDatabase(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case http.MethodGet:
		action := r.FormValue("action")

		if strings.EqualFold(action, "all_options") {
			rows, err := asql.Query(tx, "SELECT dst_db, src_db FROM syn_database")
			if err != nil {
				return nil, err
			}

			res := make([]string, 0, len(rows)*2)
			for _, row := range rows {
				res = append(res, row["dst_db"], row["src_db"])
			}

			return res, nil
		} else if strings.EqualFold(action, "src_options") {
			rows, err := asql.Query(tx, "SELECT src_db FROM syn_database")
			if err != nil {
				return nil, err
			}

			res := make([]string, 0, len(rows))
			for _, row := range rows {
				res = append(res, row["src_db"])
			}

			return res, nil
		}

		return asql.Query(tx, "SELECT * FROM syn_database")
	default:
		operation := r.PostFormValue("operation")

		// 提交数据
		id := r.PostFormValue("id")
		dstDb := strings.TrimSpace(r.PostFormValue("dst_db"))
		srcDb := strings.TrimSpace(r.PostFormValue("src_db"))
		dstFlag := strings.TrimSpace(r.PostFormValue("dst_flag"))
		srcFlag := strings.TrimSpace(r.PostFormValue("src_flag"))

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
		case "reload":
			// 重新加载目标数据库的数据库表、字段和触发器
			if len(dstDb) > 0 {
				if err := reloadDatabase(tx, dstDb); err != nil {
					return nil, err
				}
			}

			// 重新加载原始数据库的数据库表、字段和触发器
			if len(srcDb) > 0 {
				if err := reloadDatabase(tx, srcDb); err != nil {
					return nil, err
				}
			}

			return map[string]interface{}{"status": "success"}, nil
		case "compare":
			return compareTables(tx, id)
		default:
			return nil, fmt.Errorf("unexpect operation %q", operation)
		}
	}
}
