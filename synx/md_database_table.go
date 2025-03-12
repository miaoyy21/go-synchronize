package synx

import (
	"database/sql"
	"go-synchronize/asql"
	"net/http"
)

func MDDatabaseTable(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
	switch r.Method {
	case http.MethodGet:
		// 查询请求
		return asql.Query(tx, "SELECT * FROM syn_database_table ")
	default:

	}

	return nil, nil
}
