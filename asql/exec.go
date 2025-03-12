package asql

import (
	"database/sql"
	"github.com/sirupsen/logrus"
)

func Exec(tx *sql.Tx, query string, args ...interface{}) error {
	logrus.Debugf("%s %s", fnArgs(args...), query)

	_, err := tx.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}
