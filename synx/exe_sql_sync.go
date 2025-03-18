package synx

import (
	"bytes"
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
	"strings"
)

type SyncColumnPolicy struct {
	Code  string
	Name  string
	Index int
}

type SqlSyncColumn struct {
	Name      string
	Type      string
	IsPrimary bool
	IsLast    bool

	Policy SyncColumnPolicy
}

type SqlSync struct {
	SrcDatabase string
	DstDatabase string

	Table  string
	IsSync bool

	SrcFlag string
	DstFlag string
	Columns []SqlSyncColumn
}

func ExeSqlSync(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case http.MethodGet:
		database, table := r.FormValue("database_name"), r.FormValue("table_name")
		if len(database) > 1 {

			var query string
			var args []interface{}

			if len(table) > 1 {
				query = "SELECT db.dst_db AS database_name, src.table_name, src.is_sync FROM syn_src_table src INNER JOIN syn_database db ON db.src_db = src.database_name WHERE src.database_name = ? AND src.table_name = ? ORDER BY src.table_name ASC"
				args = []interface{}{database, table}
			} else {
				query = "SELECT db.dst_db AS database_name, src.table_name, src.is_sync FROM syn_src_table src INNER JOIN syn_database db ON db.src_db = src.database_name WHERE src.database_name = ? ORDER BY src.table_name ASC"
				args = []interface{}{database}
			}

			rows, err := asql.Query(tx, query, args...)
			if err != nil {
				return nil, err
			}

			buf := new(bytes.Buffer)
			for _, row := range rows {
				data, err := getSqlSync(tx, database, row["database_name"], row["table_name"], row["is_sync"] == "1")
				if err != nil {
					return nil, err
				}

				buf.WriteString(getTemplate("sync_sql.tpl", data))
			}

			return buf.String(), nil
		}

		return nil, fmt.Errorf("unknown query params")
	default:
		return nil, fmt.Errorf("unexpect operation ")
	}
}

func getSqlSync(tx *sql.Tx, srcDatabase, dstDatabase, table string, isSync bool) (*SqlSync, error) {
	data := &SqlSync{
		SrcDatabase: srcDatabase,
		DstDatabase: dstDatabase,
		Table:       table,
		IsSync:      isSync,

		Columns: make([]SqlSyncColumn, 0),
	}

	if err := asql.QueryRow(tx, "SELECT src_flag, dst_flag FROM syn_database WHERE src_db = ?", srcDatabase).Scan(&data.SrcFlag, &data.DstFlag); err != nil {
		return nil, err
	}

	cols, err := asql.Query(tx, `
		SELECT column_name, column_type, is_primary, policy_code, policy_name
		FROM (
			SELECT T.column_id, T.column_name, T.column_type, T.is_primary, X.code AS policy_code, X.name AS policy_name
			FROM syn_src_policy T
				LEFT JOIN syn_column_policy X ON X.code = T.column_policy
			WHERE T.database_name = ? AND T.table_name = ? 
			UNION 
			SELECT 9999, '_flag_', 'VARCHAR(1)', '0', 'None', '-'
		) TT
		ORDER BY column_id ASC
	`, srcDatabase, table)
	if err != nil {
		return nil, err
	}

	indexes := make(map[string]int)
	for i, col := range cols {
		policyCode, policyName := col["policy_code"], col["policy_name"]
		if !strings.EqualFold(policyCode, "None") {
			indexes[policyCode]++
		}
		column := SqlSyncColumn{
			Name:      col["column_name"],
			Type:      col["column_type"],
			IsPrimary: col["is_primary"] == "1",
			IsLast:    i+1 == len(cols),

			Policy: SyncColumnPolicy{
				Code:  policyCode,
				Name:  policyName,
				Index: indexes[policyCode],
			},
		}

		data.Columns = append(data.Columns, column)
	}

	return data, nil
}
