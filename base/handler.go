package base

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/sirupsen/logrus"
	"net/http"
	"runtime/debug"
)

func Handler(db *sql.DB, handler func(db *sql.Tx, w http.ResponseWriter, r *http.Request) (interface{}, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if msg := recover(); msg != nil {
				debug.PrintStack()
				handlerError(w, fmt.Errorf("PANIC Handler() :: %s", msg))
			}
		}()

		// 开启事务
		tx, err := db.Begin()
		if err != nil {
			handlerError(w, err)
			return
		}

		// 事务处理
		res, err := handler(tx, w, r)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				logrus.Errorf("Handler() Request Rollback Failure :: %s", err.Error())
			}

			if sqlErr, ok := err.(mssql.Error); ok {
				handlerError(w, errors.New(sqlErr.Error()))
			} else {
				handlerError(w, err)
			}

			return
		}

		if err := tx.Commit(); err != nil {
			handlerError(w, err)
			return
		}

		if s, ok := res.(string); ok {
			if _, err := w.Write([]byte(s)); err != nil {
				logrus.Errorf("Handler() Response Write Text Failure %s", err.Error())
			}
		} else {
			encode := json.NewEncoder(w)
			encode.SetIndent("", "\t")

			if err := encode.Encode(res); err != nil {
				logrus.Errorf("Handler() Response Write JSON Failure %s", err.Error())
			}
		}
	})
}

func handlerError(w http.ResponseWriter, msg error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	bs, err := json.Marshal(map[string]interface{}{"status": "error", "error": msg.Error()})
	if err != nil {
		logrus.Errorf("handlerError() :: Handler Error Failure %s", err.Error())
		return
	}

	if _, err := w.Write(bs); err != nil {
		logrus.Errorf("handlerError() :: HTTP Write Failure %s", err.Error())
	}
}
