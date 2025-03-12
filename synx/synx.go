package synx

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"go-synchronize/base"
	"path/filepath"
	"text/template"
)

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
