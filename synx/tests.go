package synx

import (
	"database/sql"
	"net/http"
)

func Tests(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
	return nil, nil
}
