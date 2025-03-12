package synx

import (
	"database/sql"
	"fmt"
	"net/http"
)

func MDDatabaseTable(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
	switch r.Method {
	case http.MethodGet:
		// 查询指定数据的所有表信息
		return loadTables(tx, r.FormValue("database"))
	default:
		operation := r.PostFormValue("operation")

		return nil, fmt.Errorf("unexpect operation %q", operation)
	}
}
