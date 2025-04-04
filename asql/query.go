package asql

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
	"unicode"
)

func Query(tx *sql.Tx, query string, args ...interface{}) ([]map[string]string, error) {
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

	// Rows
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Columns
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]sql.RawBytes, len(columns))
	valuePts := make([]interface{}, len(columns))
	for i := 0; i < len(columns); i++ {
		valuePts[i] = &values[i]
	}

	entries := make([]map[string]string, 0)
	for rows.Next() {
		if err := rows.Scan(valuePts...); err != nil {
			return nil, err
		}

		entry := make(map[string]string)
		for i, col := range columns {
			if values[i] != nil {
				value := string(values[i])

				// 隐式转换时间格式
				if len(value) == len("2006-01-02T15:04:05Z") {
					dt, err := time.Parse(time.RFC3339, value)
					if err == nil {
						if dt.Hour()+dt.Minute()+dt.Second() == 0 {
							value = dt.Format("2006-01-02")
						} else {
							value = dt.Format("2006-01-02 15:04:05")
						}
					}
				}

				entry[strings.ToLower(col)] = value
			}
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func QueryHashed(tx *sql.Tx, idField string, query string, args ...interface{}) (map[string]string, error) {
	// Rows
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Columns
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]sql.RawBytes, len(columns))
	valuePts := make([]interface{}, len(columns))
	for i := 0; i < len(columns); i++ {
		valuePts[i] = &values[i]
	}

	hashed := make(map[string]string)
	for rows.Next() {
		if err := rows.Scan(valuePts...); err != nil {
			return nil, err
		}

		full, id := make([]byte, 0), sql.RawBytes{}
		for i, col := range columns {
			if values[i] == nil {
				full = append(full, 0)
				continue
			}

			if strings.EqualFold(col, idField) {
				id = values[i]
			} else {
				full = append(full, values[i]...)
			}
		}

		if len(bytes.TrimSpace(id)) < 1 {
			return nil, fmt.Errorf("missing id field value")
		}

		h5 := md5.Sum(full)
		hashed[string(id)] = hex.EncodeToString(h5[:])
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hashed, nil
}
