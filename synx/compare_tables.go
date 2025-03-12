package synx

import (
	"database/sql"
	"go-synchronize/asql"
)

func compareTables(tx *sql.Tx, dstDb string, srcDb string) (interface{}, error) {

	return asql.Query(tx, getTemplate("compare_tables.tpl",
		struct {
			Dst string
			Src string
		}{
			Dst: dstDb,
			Src: srcDb,
		},
	))
}
