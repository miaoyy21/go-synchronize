package synx

import (
	"database/sql"
	"github.com/sirupsen/logrus"
	"go-synchronize/asql"
	"sort"
	"strings"
)

type Column struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	IsPrimary  string `json:"is_primary"`
	IsNullable string `json:"is_nullable"`
	IsIdentity string `json:"is_identity"`
}

type Table struct {
	Name     string    `json:"name"`     // 数据库表名称
	Rows     string    `json:"rows"`     // 行数
	Primary  []string  `json:"primary"`  // 主键
	Columns  []*Column `json:"columns"`  // 字段
	Triggers []string  `json:"triggers"` // 触发器【启用】
}

// 重新加载数据库的数据库表、字段和触发器
func reloadDatabase(tx *sql.Tx, database string) error {

	// 重载数据库表和字段
	if err := asql.Exec(tx, getTemplate("reload_table_columns.tpl", struct{ Database string }{database})); err != nil {
		return err
	}

	// 重载数据库表和触发器
	if err := asql.Exec(tx, getTemplate("reload_table_triggers.tpl", struct{ Database string }{database})); err != nil {
		return err
	}

	return nil
}

// 读取已加载的数据库表和触发器
func loadTables(tx *sql.Tx, database string) ([]*Table, error) {
	tables := make(map[string]*Table)

	// 查询数据库表和字段
	cols, err := asql.Query(tx, "SELECT table_name, column_name, column_type, is_primary, is_nullable, is_identity FROM syn_table_column WHERE database_name = ? ORDER BY table_name ASC,column_id ASC", database)
	if err != nil {
		return nil, err
	}

	// 处理数据库表和字段
	for _, col := range cols {
		tableName := strings.ToLower(col["table_name"])

		// Table
		table, ok := tables[tableName]
		if !ok {
			table = &Table{
				Name:     tableName,
				Columns:  make([]*Column, 0),
				Triggers: make([]string, 0),
			}
		}

		// Column
		column := &Column{
			Name:       col["column_name"],
			Type:       col["column_type"],
			IsPrimary:  col["is_primary"],
			IsNullable: col["is_nullable"],
			IsIdentity: col["is_identity"],
		}

		// Primary
		if col["is_primary"] == "1" {
			table.Primary = append(table.Primary, col["column_name"])
		}

		// Columns
		table.Columns = append(table.Columns, column)

		// Tables
		tables[tableName] = table
	}

	// 查询数据记录数量
	rows, err := asql.Query(tx, "SELECT table_name, rows FROM syn_table WHERE database_name = ?", database)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		tableName, tableRows := strings.ToLower(row["table_name"]), row["rows"]
		table, ok := tables[tableName]
		if !ok {
			logrus.Panicf("unexpect table %q", tableName)
		}

		table.Rows = tableRows
		tables[tableName] = table
	}

	// 查询数据库表的触发器
	tris, err := asql.Query(tx, "SELECT table_name, trigger_name FROM syn_table_trigger WHERE database_name = ? ORDER BY table_name ASC", database)
	if err != nil {
		return nil, err
	}

	// 处理数据库表的触发器
	for _, tri := range tris {
		tableName, triggerName := strings.ToLower(tri["table_name"]), tri["trigger_name"]
		table, ok := tables[tableName]
		if !ok {
			logrus.Panicf("unexpect trigger %q of %q", triggerName, tableName)
		}

		// Triggers
		table.Triggers = append(table.Triggers, triggerName)
		tables[tableName] = table
	}

	// All Table Name
	tableNames := make([]string, 0, len(tables))
	for tableName := range tables {
		tableNames = append(tableNames, tableName)
	}
	sort.Strings(tableNames)

	// All Table
	resTables := make([]*Table, 0, len(tables))
	for _, tableName := range tableNames {
		resTables = append(resTables, tables[tableName])
	}

	return resTables, nil
}
