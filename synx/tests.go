package synx

import (
	"database/sql"
	"go-synchronize/asql"
	"net/http"
)

func Tests(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
	cols, keys, rows, err := asql.QueryFull(tx, "id", "select * from syn_table_column")
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"cols": cols, "keys": keys, "rows": rows}, nil
}
