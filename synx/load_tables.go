package synx

import (
	"database/sql"
	"github.com/sirupsen/logrus"
	"go-synchronize/asql"
	"sort"
	"strings"
)

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Table struct {
	Name     string    `json:"name"`     // 数据库表名称
	Primary  *Column   `json:"primary"`  // 主键
	Columns  []*Column `json:"columns"`  // 字段
	Triggers []string  `json:"triggers"` // 触发器【启用】
}

func loadDatabase(tx *sql.Tx, database string) error {

	// 查询数据库表和字段
	cols, err := asql.Query(tx, getTemplate("table_columns.tpl", struct{ Database string }{database}))
	if err != nil {
		return err
	}

	// 添加数据库表和字段
	for _, col := range cols {
		if err := asql.Insert(tx,
			"INSERT INTO syn_table_column(id, database_name, table_name, column_name, column_type, is_primary, create_at) VALUES (?,?,?,?,?,?,?)",
			asql.GenerateId(), database, col["table_name"], col["column_name"], col["column_type"], col["is_primary"], asql.GetDateTime(),
		); err != nil {
			return err
		}
	}

	// 先清空数据库表和触发器
	if err := asql.Delete(tx, "DELETE FROM syn_table_trigger WHERE database_name = ?", database); err != nil {
		return err
	}

	// 查询数据库表和触发器
	tris, err := asql.Query(tx, getTemplate("table_triggers.tpl", struct{ Database string }{database}))
	if err != nil {
		return err
	}

	// 添加数据库表和触发器
	for _, tri := range tris {
		if err := asql.Insert(tx,
			"INSERT INTO syn_table_trigger(id, database_name, table_name, trigger_name, create_at) VALUES (?,?,?,?,?)",
			asql.GenerateId(), database, tri["table_name"], tri["trigger_name"], asql.GetDateTime(),
		); err != nil {
			return err
		}
	}

	return nil
}

func loadTables(tx *sql.Tx, database string) ([]*Table, error) {
	tables := make(map[string]*Table)

	// 查询数据库表和字段
	cols, err := asql.Query(tx, "SELECT * FROM syn_table_column WHERE database_name = ?", database)
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
			Name: col["column_name"],
			Type: col["column_type"],
		}

		// Primary
		if col["is_primary"] == "1" {
			table.Primary = column
		}

		// Columns
		table.Columns = append(table.Columns, column)

		// Tables
		tables[tableName] = table
	}

	// 查询数据库表的触发器
	tris, err := asql.Query(tx, "SELECT * FROM syn_table_trigger WHERE database_name = ?", database)
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
