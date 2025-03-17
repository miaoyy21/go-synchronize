package synx

import (
	"bytes"
	"database/sql"
	"fmt"
	"go-synchronize/asql"
	"net/http"
)

type DifferenceType string

var (
	DifferenceTypeCreateTable  = "Create Table"
	DifferenceTypeAddColumn    = "Add Column"
	DifferenceTypeModifyColumn = "Modify Column"
)

type TableSyncColumn struct {
	Name      string
	Type      string
	IsPrimary bool
	TypeOrg   string
}

type TableSync struct {
	Database string
	Table    string
	IsSync   bool

	CreateTable  []TableSyncColumn
	AddColumn    []TableSyncColumn
	ModifyColumn []TableSyncColumn
}

func ExeTableSync(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case http.MethodGet:
		database, table := r.FormValue("database_name"), r.FormValue("table_name")
		if len(database) > 1 {
			buf := new(bytes.Buffer)

			var query string
			var args []interface{}

			if len(table) > 1 {
				query = "SELECT table_name, is_sync FROM syn_src_table WHERE database_name = ? AND table_name = ? ORDER BY table_name ASC"
				args = []interface{}{database, table}
			} else {
				query = "SELECT table_name, is_sync FROM syn_src_table WHERE database_name = ? ORDER BY table_name ASC"
				args = []interface{}{database}
			}

			rows, err := asql.Query(tx, query, args...)
			if err != nil {
				return nil, err
			}

			for _, row := range rows {
				data, err := getTableSync(tx, database, row["table_name"], row["is_sync"] == "1")
				if err != nil {
					return nil, err
				}

				buf.WriteString(getTemplate("sync_table.tpl", data))
			}

			return buf.String(), nil
		}

		return nil, fmt.Errorf("unknown query params")
	default:
		return nil, fmt.Errorf("unexpect operation ")
	}
}

func getTableSync(tx *sql.Tx, database, table string, isSync bool) (*TableSync, error) {
	data := &TableSync{
		Database: database,
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
	`, database, table)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		column := TableSyncColumn{
			Name:      row["column_name"],
			Type:      row["column_type"],
			IsPrimary: row["is_primary"] == "1",
			TypeOrg:   row["column_type_org"],
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

	return data, nil
}
