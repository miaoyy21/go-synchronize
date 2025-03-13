package synx

import (
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
)

func MDSynSrcPolicy(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
	switch r.Method {
	case http.MethodGet:
		return asql.Query(tx, "SELECT * FROM syn_src_table")
	default:
		operation := r.PostFormValue("operation")

		return nil, fmt.Errorf("unexpect operation %q", operation)
	}
}
