package main

import (
	"database/sql"
	"github.com/antonfisher/nested-logrus-formatter"
	_ "github.com/denisenkom/go-mssqldb" // SQL Server 驱动
	"github.com/sirupsen/logrus"
	"go-synchronize/base"
	"go-synchronize/synx"
	"net"
	"net/http"
	"os"
)

func main() {
	// 默认的日志级别
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetFormatter(&formatter.Formatter{TimestampFormat: "2006-01-02 15:04:05", HideKeys: true})

	// 获取当前目录
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatalf("os.Getwd() Failure :: %s", err.Error())
	}

	// 初始化
	if err := base.Init(dir); err != nil {
		logrus.Fatalf("synx.Init() Failure :: %s", err.Error())
	}

	// 数据库链接
	db, err := sql.Open("mssql", base.Config.DataSource)
	if err != nil {
		logrus.Fatalf("sql.Open() Failure :: %s", err.Error())
	}

	// Ping ...
	if err := db.Ping(); err != nil {
		logrus.Fatalf("db.Ping() Failure :: %s", err.Error())
	}
	logrus.Info("连接数据库成功 ...")

	// 静态文件
	http.Handle("/", http.FileServer(http.Dir("www")))

	// 访问服务
	http.Handle("/api/tests", base.Handler(db, synx.Tests))
	http.Handle("/api/md/database", base.Handler(db, synx.MDDatabase))
	http.Handle("/api/md/database_table", base.Handler(db, synx.MDDatabaseTable))
	http.Handle("/api/md/column_rule", base.Handler(db, synx.MDColumnRule))

	http.Handle("/api/md/column_policy", base.Handler(db, synx.MDColumnPolicy))
	http.Handle("/api/syn/src_table", base.Handler(db, synx.MDSynSrcTable))
	http.Handle("/api/syn/src_policy", base.Handler(db, synx.MDSynSrcPolicy))

	addr := net.JoinHostPort(base.Config.Host, base.Config.Port)
	logrus.Infof("HTTP服务器监听地址: %s ......", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logrus.Errorf("Listen Failure %s", err.Error())
	}
}
