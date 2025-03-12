package synx

import (
	"database/sql"
	"go-synchronize/asql"
)

func compareTables(tx *sql.Tx, dstDb string, srcDb string) (interface{}, error) {

	if err := asql.Exec(tx, getTemplate("compare_tables.tpl",
		struct {
			DstDb string
			SrcDb string
		}{
			DstDb: dstDb,
			SrcDb: srcDb,
		},
	)); err != nil {
		return nil, err
	}

	return nil, nil
}
