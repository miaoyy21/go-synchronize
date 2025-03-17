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

	for {
		select {
		case <-ticker.C:
			logrus.Debugf("执行定时任务 ... ")

			tx, err := db.Begin()
			if err != nil {
				logrus.Errorf("db.Begin() failure :: %s", err.Error())
				continue
			}

			var id, srcDriver, srcDatasource, srcSql, srcIdField string
			var dstDriver, dstDatasource, dstSql, dstTable, dstIdField string

			query := `
			SELECT TA.id, 
				M1.driver AS src_driver, M1.datasource AS src_datasource, TA.src_sql, TA.src_id_field,
				M2.driver AS dst_driver, M2.datasource AS dst_datasource, TA.dst_sql, TA.dst_table, TA.dst_id_field
			FROM syn_datasource_sync TA 
				INNER JOIN syn_datasource M1 ON TA.src_ds_code = M1.code
				INNER JOIN syn_datasource M2 ON TA.dst_ds_code = M2.code
			WHERE NOT EXISTS (SELECT 1 FROM syn_datasource_sync TB WHERE TB.sync_status = ?)
				AND TA.sync_status = ?
			ORDER BY TA.sync_at ASC,TA.order_ ASC
			`

			if err := asql.QueryRow(tx, query, SyncStatusExecuting, SyncStatusWaiting).
				Scan(&id, &srcDriver, &srcDatasource, &srcSql, &srcIdField, &dstDriver, &dstDatasource, &dstSql, &dstTable, &dstIdField); err != nil {
				if err == sql.ErrNoRows {
					logrus.Debug("<<< none tasks found >>>")
					continue
				}

				logrus.Errorf("asql.QueryRow() failure :: %s", err.Error())
				continue
			}

			if err := asql.Update(tx, "UPDATE syn_datasource_sync SET sync_status = ? WHERE id = ?", SyncStatusExecuting, id); err != nil {
				logrus.Errorf("asql.Update() failure :: %s", err.Error())
				_ = tx.Rollback()
				continue
			}

			if err := run(srcDriver, srcDatasource, srcSql, srcIdField, dstDriver, dstDatasource, dstSql, dstTable, dstIdField); err != nil {
				logrus.Errorf("run() failure :: %s", err.Error())
				_ = tx.Rollback()
				continue
			}

			if err := asql.Update(tx, "UPDATE syn_datasource_sync SET sync_status = ?, sync_at = ? WHERE id = ?", SyncStatusWaiting, asql.GetDateTime(), id); err != nil {
				logrus.Errorf("asql.Update() failure :: %s", err.Error())
				_ = tx.Rollback()
				continue
			}

			_ = tx.Commit()
		}
	}
}

func run(srcDriver, srcDatasource, srcSql, srcIdField string, dstDriver, dstDatasource, dstSql, dstTable, dstIdField string) error {
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

	// src cols && rows
	_, srcHashed, srcRows, err := asql.QueryHashed(srcTx, srcIdField, srcSql)
	if err != nil {
		_, _ = srcTx.Rollback(), dstTx.Rollback()
		return err
	}

	// dst cols && rows
	dstCols, dstHashed, _, err := asql.QueryHashed(dstTx, dstIdField, dstSql)
	if err != nil {
		_, _ = srcTx.Rollback(), dstTx.Rollback()
		return err
	}

	// compare src && dst Map
	added, changed, removed := base.CompareMap(dstHashed, srcHashed)

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

		if err := asql.Insert(dstTx, query, args...); err != nil {
			return err
		}
	}

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
		if err := asql.Update(dstTx, query, args...); err != nil {
			return err
		}
	}

	// Removed
	for key := range removed {
		query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", dstTable, dstIdField)
		if err := asql.Delete(dstTx, query, key); err != nil {
			return err
		}
	}

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
