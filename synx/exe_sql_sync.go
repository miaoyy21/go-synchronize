package synx

import (
	"bytes"
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
	"sort"
)

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

			allData := make([]*TableSync, 0)
			for _, row := range rows {
				data, err := getSqlSync(tx, database, row["database_name"], row["table_name"], row["is_sync"] == "1")
				if err != nil {
					return nil, err
				}

				allData = append(allData, data)
			}

			// 排序：创建表优先，否则会存在一起依赖有问题
			sort.Slice(allData, func(i, j int) bool {
				return len(allData[i].CreateTable) > 0
			})

			buf := new(bytes.Buffer)
			for _, data := range allData {
				buf.WriteString(getTemplate("sync_table.tpl", data))
			}

			return buf.String(), nil
		}

		return nil, fmt.Errorf("unknown query params")
	default:
		return nil, fmt.Errorf("unexpect operation ")
	}
}

func getSqlSync(tx *sql.Tx, srcDatabase, dstDatabase, table string, isSync bool) (*TableSync, error) {
	data := &TableSync{
		Database: dstDatabase,
		Table:    table,
		IsSync:   isSync,

		CreateTable:  make([]TableSyncColumn, 0),
		AddColumn:    make([]TableSyncColumn, 0),
		ModifyColumn: make([]TableSyncColumn, 0),
	}

	rows, err := asql.Query(tx, `
		SELECT difference_type, column_name, column_type, is_primary, column_type_org 
		FROM syn_src_difference 
		WHERE database_name = ? AND table_name = ? 
		ORDER BY column_id ASC 
	`, srcDatabase, table)
	if err != nil {
		return nil, err
	}

	if err := asql.QueryRow(tx, `
		SELECT CASE WHEN NOT EXISTS (SELECT 1 FROM syn_table_column dst WHERE dst.database_name = db.dst_db AND dst.table_name = ? AND dst.column_name = ?) THEN db.dst_flag ELSE '' END AS flag
		FROM syn_database db
		WHERE db.dst_db = ?
	`, table, "_flag_", dstDatabase).Scan(&data.Flag); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	for _, row := range rows {
		column := TableSyncColumn{
			Name:      row["column_name"],
			Type:      row["column_type"],
			IsPrimary: row["is_primary"] == "1",
			TypeOrg:   row["column_type_org"],
		}

		if column.IsPrimary {
			data.Primary = column
		}

		switch row["difference_type"] {
		case DifferenceTypeCreateTable:
			data.CreateTable = append(data.CreateTable, column)
		case DifferenceTypeAddColumn:
			data.AddColumn = append(data.AddColumn, column)
		case DifferenceTypeModifyColumn:
			data.ModifyColumn = append(data.ModifyColumn, column)
		default:
			return nil, fmt.Errorf("invalid difference type of %q", row["difference_type"])
		}
	}

	if len(data.CreateTable) > 0 {
		data.Flag = ""
	}

	return data, nil
}
