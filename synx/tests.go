package synx

import (
	"database/sql"
	"net/http"
)

func Tests(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
	database := "jz24_scxt"

	// 加载数据库的数据库表、字段和触发器
	if err := reloadDatabase(tx, database); err != nil {
		return nil, err
	}

	return loadTables(tx, database)
}
