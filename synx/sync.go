package synx

import (
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb" // SQL Server 驱动
	"github.com/sirupsen/logrus"
	"go-synchronize/asql"
	"go-synchronize/base"
	"strings"
	"time"
)

func Run(db *sql.DB) {
	ticker := time.NewTicker(15 * time.Second)

	_, _ = db.Exec("UPDATE syn_datasource_sync SET sync_status = ? WHERE sync_status = ?", SyncStatusWaiting, SyncStatusExecuting)

	for {
		select {
		case <-ticker.C:
			logrus.Debugf("执行定时任务 ... ")

			var id, srcDriver, srcDatasource, srcSql string
			var dstDriver, dstDatasource, dstSql, dstTable, dstIdField string

			query := "SELECT TA.id, \n" +
				"	M1.driver AS src_driver, M1.datasource AS src_datasource, TA.src_sql, \n" +
				"	M2.driver AS dst_driver, M2.datasource AS dst_datasource, TA.dst_sql, TA.dst_table, TA.dst_id_field \n" +
				"FROM syn_datasource_sync TA \n" +
				"	INNER JOIN syn_datasource M1 ON TA.src_ds_code = M1.code \n" +
				"	INNER JOIN syn_datasource M2 ON TA.dst_ds_code = M2.code \n" +
				"WHERE NOT EXISTS (SELECT 1 FROM syn_datasource_sync TB WHERE TB.sync_status = ?) \n" +
				"	AND TA.sync_status = ? \n" +
				"ORDER BY TA.sync_at ASC,TA.order_ ASC"

			if err := db.QueryRow(query, SyncStatusExecuting, SyncStatusWaiting).
				Scan(&id, &srcDriver, &srcDatasource, &srcSql, &dstDriver, &dstDatasource, &dstSql, &dstTable, &dstIdField); err != nil {
				if err == sql.ErrNoRows {
					logrus.Debug("<<< empty sync tasks found >>>")
					continue
				}

				logrus.Errorf("asql.QueryRow() failure :: %s", err.Error())
				continue
			}

			if _, err := db.Exec("UPDATE syn_datasource_sync SET sync_status = ? WHERE id = ?", SyncStatusExecuting, id); err != nil {
				logrus.Errorf("asql.Update() failure :: %s", err.Error())
				continue
			}

			if err := run(srcDriver, srcDatasource, srcSql, dstDriver, dstDatasource, dstSql, dstTable, dstIdField); err != nil {
				logrus.Errorf("run() failure :: %s", err.Error())

				if _, err := db.Exec("UPDATE syn_datasource_sync SET sync_status = ? WHERE id = ? AND sync_status = ? ", SyncStatusWaiting, id, SyncStatusExecuting); err != nil {
					logrus.Errorf("asql.Update() failure :: %s", err.Error())
				}

				continue
			}

			if _, err := db.Exec("UPDATE syn_datasource_sync SET sync_status = ?, sync_at = ? WHERE id = ? AND sync_status = ?", SyncStatusWaiting, asql.GetDateTime(), id, SyncStatusExecuting); err != nil {
				logrus.Errorf("asql.Update() failure :: %s", err.Error())
				continue
			}
		}
	}
}

func run(srcDriver, srcDatasource, srcSql string, dstDriver, dstDatasource, dstSql, dstTable, dstIdField string) error {
	// src
	srcTx, err := initDB(srcDriver, srcDatasource)
	if err != nil {
		return err
	}

	// dst
	dstTx, err := initDB(dstDriver, dstDatasource)
	if err != nil {
		_ = srcTx.Rollback()
		return err
	}

	// transaction Rollback
	rollback := func() {
		_, _ = srcTx.Rollback(), dstTx.Rollback()
	}

	// src cols && rows
	_, srcHashed, srcRows, err := asql.QueryHashed(srcTx, dstIdField, srcSql)
	if err != nil {
		rollback()
		return err
	}
	logrus.Debugf("source %d rows .", len(srcHashed))

	// dst cols && rows
	dstCols, dstHashed, _, err := asql.QueryHashed(dstTx, dstIdField, dstSql)
	if err != nil {
		rollback()
		return err
	}
	logrus.Debugf("target %d rows .", len(dstHashed))

	// compare src && dst Map
	added, changed, removed := base.CompareMap(dstHashed, srcHashed)
	logrus.Debugf("compare :: added %d rows, changed %d rows, and removed %d rows .", len(added), len(changed), len(removed))
	if len(added)+len(changed)+len(removed) < 1 {
		return nil
	}

	// Added
	for key := range added {
		values := srcRows[key]

		query := fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s)", dstTable, strings.Join(dstCols, ","), strings.TrimRight(strings.Repeat("?,", len(dstCols)), ","))

		args := make([]interface{}, 0, len(dstCols))
		for _, col := range dstCols {
			value, ok := values[col]
			if !ok {
				args = append(args, nil)
				continue
			}

			args = append(args, value)
		}

		if _, err := dstTx.Exec(query, args...); err != nil {
			rollback()
			return err
		}
	}
	logrus.Debugf("%d rows added ...", len(added))

	// Changed
	for key := range changed {
		values, args := srcRows[key], make([]interface{}, 0, len(dstCols))

		for _, col := range dstCols {
			value, ok := values[col]
			if !ok {
				args = append(args, nil)
				continue
			}

			args = append(args, value)
		}
		args = append(args, key)

		query := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", dstTable, strings.Join(dstCols, " = ?,"), dstIdField)
		if _, err := dstTx.Exec(query, args...); err != nil {
			rollback()
			return err
		}
	}
	logrus.Debugf("%d rows changed ...", len(added))

	// Removed
	for key := range removed {
		query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", dstTable, dstIdField)
		if _, err := dstTx.Exec(query, key); err != nil {
			rollback()
			return err
		}
	}
	logrus.Debugf("%d rows removed ...", len(added))

	return dstTx.Commit()
}

func initDB(driver string, datasource string) (*sql.Tx, error) {

	// 数据库链接
	db, err := sql.Open(driver, datasource)
	if err != nil {
		logrus.Fatalf("sql.Open() Failure :: %s", err.Error())
	}

	// Ping ...
	if err := db.Ping(); err != nil {
		logrus.Fatalf("db.Ping() Failure :: %s", err.Error())
	}

	return db.Begin()
}
