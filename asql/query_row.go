package asql

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"unicode"
)

func QueryRow(tx *sql.Tx, query string, args ...interface{}) *sql.Row {
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		var prefix string

		// 格式化SQL输出
		for i := 5; i > 1; i-- {
			prefix = fmt.Sprintf("\n%s", strings.Repeat("\t", i))
			if strings.HasPrefix(query, prefix) {
				break
			}
		}
		query = strings.ReplaceAll(query, prefix, "\n\t")

		query = strings.TrimRightFunc(query, unicode.IsSpace)
		logrus.Debugf("%s %s", fnArgs(args...), query)
	}

	return tx.QueryRow(query, args...)
}
