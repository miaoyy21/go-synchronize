package synx

import (
	"database/sql"
	"go-synchronize/asql"
	"net/http"
)

func Tests(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
	id := r.FormValue("id")

	var dstDb, srcDb string

	if err := asql.QueryRow(tx, "SELECT dst_db, src_db FROM syn_database WHERE id = ?", id).Scan(&dstDb, &srcDb); err != nil {
		return nil, err
	}

	return compareTables(tx, dstDb, srcDb)
}
