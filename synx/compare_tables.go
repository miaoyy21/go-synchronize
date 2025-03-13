package synx

import (
	"database/sql"
	"go-synchronize/asql"
)

func compareTables(tx *sql.Tx, id string) (interface{}, error) {
	var dstDb, srcDb string

	if err := asql.QueryRow(tx, "SELECT dst_db, src_db FROM syn_database WHERE id = ?", id).Scan(&dstDb, &srcDb); err != nil {
		return nil, err
	}

	return asql.Query(tx, getTemplate("compare_tables.tpl",
		struct {
			Dst string
			Src string
		}{Dst: dstDb, Src: srcDb},
	))
}
