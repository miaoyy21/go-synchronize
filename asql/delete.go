package asql

import (
	"database/sql"
	"github.com/sirupsen/logrus"
)

func Delete(tx *sql.Tx, query string, args ...interface{}) error {
	logrus.Debugf("%s %s", fnArgs(args...), query)

	if _, err := tx.Exec(query, args...); err != nil {
		return err
	}

	return nil
}
