package synx

import (
	"bytes"
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
)

type SyncColumnPolicy struct {
	Code string
	Name string

	Index int

	ReplaceCode    string
	IsExactlyMatch bool
}

type SqlSyncColumn struct {
	Name       string
	Type       string
	IsPrimary  bool
	IsNullable bool
	IsIdentity bool
	IsLast     bool

	Policy SyncColumnPolicy
}

type SqlSync struct {
	SrcDatabase string
	DstDatabase string

	Table       string
	IsSync      bool
	HasIdentity bool

	SrcFlag  string
	DstFlag  string
	Columns  []SqlSyncColumn
	Triggers []string
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

			buf.WriteString(fmt.Sprintf("USE %s;\n", rows[0]["database_name"]))
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
		HasIdentity: false,

		Columns:  make([]SqlSyncColumn, 0),
		Triggers: make([]string, 0),
	}

	if err := asql.QueryRow(tx, "SELECT src_flag, dst_flag FROM syn_database WHERE src_db = ?", srcDatabase).Scan(&data.SrcFlag, &data.DstFlag); err != nil {
		return nil, err
	}

	cols, err := asql.Query(tx, `
		SELECT TT.column_name, TT.column_type, TT.is_primary, XX.is_nullable, XX.is_identity, TT.policy_code, TT.policy_name, TT.replace_code, TT.is_exactly_match
		FROM (
			SELECT T.column_id, T.column_name, T.column_type, T.is_primary, X.code AS policy_code, X.name AS policy_name, X.replace_code, X.is_exactly_match
			FROM syn_src_policy T
				LEFT JOIN syn_column_policy X ON X.code = T.column_policy
			WHERE T.database_name = ? AND T.table_name = ? AND T.column_name <> '_flag_'
			UNION 
			SELECT 9999, '_flag_', 'VARCHAR(1)', '0', T.code, T.name, T.replace_code, T.is_exactly_match
			FROM syn_column_policy T
			WHERE T.code = ?
		) TT
			LEFT JOIN syn_table_column XX ON XX.database_name = ? AND XX.table_name = ? AND XX.column_name = TT.column_name
		ORDER BY TT.column_id ASC
	`, srcDatabase, table, "None", srcDatabase, table)
	if err != nil {
		return nil, err
	}

	policyIndex := 0
	for i, col := range cols {
		policyCode, policyName := col["policy_code"], col["policy_name"]

		// Policy Index
		if len(col["replace_code"]) > 0 && col["is_exactly_match"] == "1" {
			policyIndex++
		}

		// Column Policy
		policy := SyncColumnPolicy{
			Code: policyCode,
			Name: policyName,

			Index:          policyIndex,
			ReplaceCode:    col["replace_code"],
			IsExactlyMatch: col["is_exactly_match"] == "1",
		}

		// Column
		column := SqlSyncColumn{
			Name:       col["column_name"],
			Type:       col["column_type"],
			IsPrimary:  col["is_primary"] == "1",
			IsNullable: col["is_nullable"] == "1",
			IsIdentity: col["is_identity"] == "1",
			IsLast:     i+1 == len(cols),

			Policy: policy,
		}

		// Identity
		if column.IsIdentity {
			data.HasIdentity = true
		}

		data.Columns = append(data.Columns, column)
	}

	triggers, err := asql.Query(tx, "SELECT trigger_name FROM syn_table_trigger WHERE database_name = ? AND table_name = ?", dstDatabase, table)
	if err != nil {
		return nil, err
	}

	for _, trigger := range triggers {
		data.Triggers = append(data.Triggers, trigger["trigger_name"])
	}

	return data, nil
}
