package synx

import (
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
)

type SyncStatus string

const (
	SyncStatusStopped   = "Stopped"
	SyncStatusWaiting   = "Waiting"
	SyncStatusExecuting = "Executing"
)

func MDDatasourceSync(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case http.MethodGet:
		return asql.Query(tx, "SELECT * FROM syn_datasource_sync ORDER BY order_ ASC")
	default:
		operation := r.PostFormValue("operation")

		// 提交数据
		id := r.PostFormValue("id")
		srcDsCode := r.PostFormValue("src_ds_code")
		srcSql := r.PostFormValue("src_sql")
		srcIdField := r.PostFormValue("src_id_field")
		dstDsCode := r.PostFormValue("dst_ds_code")
		dstSql := r.PostFormValue("dst_sql")
		dstTable := r.PostFormValue("dst_table")
		dstIdField := r.PostFormValue("dst_id_field")

		moveId := r.PostFormValue("webix_move_id")
		moveIndex := r.PostFormValue("webix_move_index")
		moveParent := r.PostFormValue("webix_move_parent")

		switch operation {
		case "insert":
			newId, syncStatus, at := asql.GenerateId(), SyncStatusStopped, asql.GetDateTime()

			query := "INSERT INTO syn_datasource_sync(id, src_ds_code, src_sql, src_id_field, dst_ds_code, dst_sql, dst_table, dst_id_field, sync_status, order_, create_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)"
			args := []interface{}{newId, srcDsCode, srcSql, srcIdField, dstDsCode, dstSql, dstTable, dstIdField, syncStatus, asql.GenerateOrderId(), at}
			if err := asql.Insert(tx, query, args...); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "newid": newId, "sync_status": syncStatus, "create_at": at}, nil
		case "update":
			syncStatus := r.PostFormValue("sync_status")

			query := "UPDATE syn_datasource_sync SET src_ds_code = ?, src_sql = ?, src_id_field = ?, dst_ds_code = ?, dst_sql = ?, dst_table = ?, dst_id_field = ?,sync_status = ? WHERE id = ?"
			args := []interface{}{srcDsCode, srcSql, srcIdField, dstDsCode, dstSql, dstTable, dstIdField, syncStatus, id}
			if err := asql.Update(tx, query, args...); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success", "id": id}, nil
		case "delete":
			if err := asql.Delete(tx, "DELETE FROM syn_datasource_sync WHERE id = ?", id); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success"}, nil
		case "order":
			if err := asql.Order(tx, "syn_datasource_sync", id, moveId, moveIndex, moveParent); err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "success"}, nil
		default:
			return nil, fmt.Errorf("unexpect operation %q", operation)
		}
	}
}
