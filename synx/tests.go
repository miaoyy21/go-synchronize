package synx

import (
	"bytes"
	"database/sql"
	"github.com/sirupsen/logrus"
	"go-synchronize/base"
	"net/http"
	"path/filepath"
	"text/template"
)

func Tests(tx *sql.Tx, w http.ResponseWriter, r *http.Request) (res interface{}, err error) {
	buf := new(bytes.Buffer)

	tpl := template.Must(template.ParseFiles(filepath.Join(base.Config.Dir, "tpl", "table_columns.tpl")))
	tpl.Execute(buf, struct{ Database string }{"xxxxxxxxxxx"})

	logrus.Info(buf.String())

	return nil, nil
}
