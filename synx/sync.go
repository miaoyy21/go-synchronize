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
			run(db)
		}
	}
}

func run(db *sql.DB) {
	var id, srcDriver, srcDatasource, srcSql string
	var dstDriver, dstDatasource, dstSql, dstTable, dstIdField, dstCompareFields string

	defer func() {
		if _, err := db.Exec("UPDATE syn_datasource_sync SET sync_status = ? WHERE id = ? AND sync_status = ? ", SyncStatusWaiting, id, SyncStatusExecuting); err != nil {
			logrus.Errorf("asql.Update() failure :: %s", err.Error())
		}

		if err := recover(); err != nil {
			logrus.Errorf("panic :: %#v", err)
		}
	}()

	logrus.Debugf("执行定时任务 ... ")

	query := "SELECT TA.id, \n" +
		"	M1.driver AS src_driver, M1.datasource AS src_datasource, TA.src_sql, \n" +
		"	M2.driver AS dst_driver, M2.datasource AS dst_datasource, TA.dst_sql, \n" +
		"	TA.dst_table, TA.dst_id_field, TA.dst_compare_fields \n" +
		"FROM syn_datasource_sync TA \n" +
		"	INNER JOIN syn_datasource M1 ON TA.src_ds_code = M1.code \n" +
		"	INNER JOIN syn_datasource M2 ON TA.dst_ds_code = M2.code \n" +
		"WHERE NOT EXISTS (SELECT 1 FROM syn_datasource_sync TB WHERE TB.sync_status = ?) \n" +
		"	AND TA.sync_status = ? \n" +
		"ORDER BY TA.sync_at ASC,TA.order_ ASC"

	if err := db.QueryRow(query, SyncStatusExecuting, SyncStatusWaiting).
		Scan(&id, &srcDriver, &srcDatasource, &srcSql, &dstDriver, &dstDatasource, &dstSql, &dstTable, &dstIdField, &dstCompareFields); err != nil {
		if err == sql.ErrNoRows {
			return
		}

		logrus.Errorf("asql.QueryRow() failure :: %s", err.Error())
		return
	}

	if _, err := db.Exec("UPDATE syn_datasource_sync SET sync_status = ? WHERE id = ?", SyncStatusExecuting, id); err != nil {
		logrus.Errorf("asql.Update() failure :: %s", err.Error())
		return
	}

	if err := runSync(srcDriver, srcDatasource, srcSql, dstDriver, dstDatasource, dstSql, dstTable, dstIdField, dstCompareFields); err != nil {
		logrus.Errorf("runSync() failure :: %s", err.Error())

		return
	}

	if _, err := db.Exec("UPDATE syn_datasource_sync SET sync_at = ? WHERE id = ? AND sync_status = ?", asql.GetDateTime(), id, SyncStatusExecuting); err != nil {
		logrus.Errorf("asql.Update() failure :: %s", err.Error())
		return
	}
}

func runSync(srcDriver, srcDatasource, srcSql string, dstDriver, dstDatasource, dstSql, dstTable, dstIdField, dstCompareFields string) error {
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
	srcHashed, err := asql.QueryHashed(srcTx, dstIdField, getCompareSql(srcSql, dstIdField, dstCompareFields))
	if err != nil {
		rollback()
		return err
	}

	// dst cols && rows
	dstHashed, err := asql.QueryHashed(dstTx, dstIdField, getCompareSql(dstSql, dstIdField, dstCompareFields))
	if err != nil {
		rollback()
		return err
	}
	logrus.Debugf("query :: source %d rows, and target %d rows .", len(srcHashed), len(dstHashed))

	// Insert or Update Fields
	dstFields := getSqlFields(dstSql)

	// compare src && dst Map
	added, changed, removed := base.CompareMap(dstHashed, srcHashed)
	logrus.Debugf("compare :: added %d rows, changed %d rows, and removed %d rows .", len(added), len(changed), len(removed))
	if len(added)+len(changed)+len(removed) < 1 {
		return nil
	}

	// Added
	for key := range added {
		values, err := asql.Query(srcTx, getWhereSql(srcSql, dstIdField), key)
		if err != nil {
			return err
		} else if len(values) != 1 {
			return fmt.Errorf("unexpect %d rows", len(values))
		}

		query := fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s)", dstTable, strings.Join(dstFields, ","), strings.TrimRight(strings.Repeat("?,", len(dstFields)), ","))

		args := make([]interface{}, 0, len(dstFields))
		for _, col := range dstFields {
			value, ok := values[0][col]
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

		logrus.Debugf("target table %q with %q inserted ...", dstTable, key)
	}

	// Changed
	for key := range changed {
		values, err := asql.Query(srcTx, getWhereSql(srcSql, dstIdField), key)
		if err != nil {
			return err
		} else if len(values) != 1 {
			return fmt.Errorf("unexpect %d rows", len(values))
		}

		args := make([]interface{}, 0, len(dstFields))

		for _, col := range dstFields {
			value, ok := values[0][col]
			if !ok {
				args = append(args, nil)
				continue
			}

			args = append(args, value)
		}
		args = append(args, key)

		query := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", dstTable, strings.Join(dstFields, " = ?,"), dstIdField)
		if _, err := dstTx.Exec(query, args...); err != nil {
			rollback()
			return err
		}

		logrus.Debugf("target table %q with %q updated ...", dstTable, key)
	}

	// Removed
	for key := range removed {
		query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", dstTable, dstIdField)
		if _, err := dstTx.Exec(query, key); err != nil {
			rollback()
			return err
		}

		logrus.Debugf("target table %q with %q deleted ...", dstTable, key)
	}

	return dstTx.Commit()
}

func initDB(driver string, datasource string) (*sql.Tx, error) {

	// 数据库链接
	db, err := sql.Open(driver, datasource)
	if err != nil {
		panic(fmt.Sprintf("sql.Open(%q,%q) fail :: %s", driver, datasource, err.Error()))
	}

	// Ping ...
	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("db.Ping() fail :: %s", err.Error()))
	}

	return db.Begin()
}

func getSqlFields(originalSql string) []string {
	originalSql = strings.TrimSpace(originalSql)
	start := strings.Index(strings.ToLower(originalSql), "select ")
	if start < 0 {
		panic(fmt.Sprintf("invalid SQL Statement %q", originalSql))
	}

	end := strings.Index(strings.ToLower(originalSql), " from ")
	if end < 1 {
		panic(fmt.Sprintf("invalid SQL Statement %q", originalSql))
	}

	if strings.EqualFold(strings.TrimSpace(originalSql[7:end]), "*") {
		panic(fmt.Sprintf("invalid SQL Statement %q with * fields", originalSql))
	}

	oFields := strings.Split(originalSql[7:end], ",")
	nFields := make([]string, 0, len(oFields))
	for _, oField := range oFields {
		nFields = append(nFields, strings.TrimSpace(oField))
	}

	return nFields
}

func getCompareSql(originalSql string, dstIdField, compareFields string) string {
	originalSql = strings.TrimSpace(originalSql)
	start := strings.Index(strings.ToLower(originalSql), "select ")
	if start < 0 {
		panic(fmt.Sprintf("invalid SQL Statement %q", originalSql))
	}

	end := strings.Index(strings.ToLower(originalSql), " from ")
	if end < 1 {
		panic(fmt.Sprintf("invalid SQL Statement %q", originalSql))
	}

	oFields := strings.Split(compareFields, ",")
	nFields := make([]string, 0, len(oFields))

	nFields = append(nFields, dstIdField)
	for _, oField := range oFields {
		nFields = append(nFields, strings.TrimSpace(oField))
	}

	return fmt.Sprintf("%s %s %s", originalSql[:6], strings.Join(nFields, ", "), originalSql[end:])
}

func getWhereSql(originalSql string, dstIdField string) string {
	originalSql = strings.TrimSpace(originalSql)
	where := strings.Index(strings.ToLower(originalSql), " where ")
	if where < 0 {
		return fmt.Sprintf("%s WHERE %s = ?", originalSql, dstIdField)
	}

	return fmt.Sprintf("%s WHERE %s = ? AND (%s)", originalSql[:where], dstIdField, originalSql[where+7:])
}
