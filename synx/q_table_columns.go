package synx

import (
	"bytes"
	"database/sql"
	"github.com/sirupsen/logrus"
	"go-synchronize/asql"
	"go-synchronize/base"
	"path/filepath"
	"text/template"
)

type Column struct {
	Name string
	Type string
}

type Table struct {
	Name     string   // 名称
	Primary  Column   // 主键
	Columns  []Column // 字段
	Triggers []string // 已启用的触发器
}

func qTable(tx *sql.Tx, database string) ([]Table, error) {

	// 查询表和字段
	columns, err := asql.Query(tx, getTemplate("table_columns.tpl", struct{ Database string }{"code"}))
	if err != nil {
		return nil, err
	}

	// 查询表和触发器
	triggers, err := asql.Query(tx, getTemplate("table_triggers.tpl", struct{ Database string }{"code"}))
	if err != nil {
		return nil, err
	}

	_ = columns
	_ = triggers

	return nil, nil
}

func getTemplate(name string, data interface{}) string {
	buf := new(bytes.Buffer)

	// 执行模版
	tpl := template.Must(template.ParseFiles(filepath.Join(base.Config.Dir, "tpl", name)))
	if err := tpl.Execute(buf, data); err != nil {
		logrus.Panicf("template Execute %q with %#v PANIC :: %s", name, data, err.Error())
	}

	// 获取待执行的SQL语句
	return buf.String()
}
