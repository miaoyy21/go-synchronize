package synx

import (
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
)

func ExeSqlSync(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case http.MethodGet:
		return asql.Query(tx, "SELECT * FROM syn_datasource_sync ORDER BY order_ ASC")
	default:
		return nil, fmt.Errorf("unexpect operation ")
	}
}
