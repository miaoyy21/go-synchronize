package synx

import (
	"database/sql"
	"net/http"
)

func Tests(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
	//cols, hashed, rows, err := asql.QueryHashed(tx, "id", "select * from syn_table_column")
	//if err != nil {
	//	return nil, err
	//}
	//
	//return map[string]interface{}{"cols": cols, "hashed": hashed, "rows": rows}, nil

	return nil, nil
}
