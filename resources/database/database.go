package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-sql-driver/mysql" // we just need it here due design
	"github.com/jmoiron/sqlx"
	"godep.lzd.co/metrics"
	"godep.lzd.co/metrics/mysqlmon"
	"godep.lzd.co/service/interfaces"
)

const (
	driver       = "mysql"
	metricPrefix = "mysql"
)

type status struct {
	Header []string   `json:"header"`
	Data   [][]string `json:"data"`
}

// DbAdapter keeps DB
type DbAdapter struct {
	dbs   []*MonitoringWrapper
	count uint64 // Monotonically incrementing counter on each query

	debug bool
}

func (db *DbAdapter) Caption() string {
	return "Database"
}

func (db *DbAdapter) Status() interfaces.Status {
	rows := make([]interfaces.StatusRow, 0, len(db.dbs))
	for i := range db.dbs {
		rows = append(rows, db.dbs[i].getStatusRow())
	}
	return interfaces.Status{
		Header: []string{
			"DB",
			"Connections",
			"Requests",
			"Errors",
			"Time avg (ms)",
		},
		Rows: rows,
	}
}

type MonitoringWrapper struct {
	Db   *sqlx.DB
	Conf *mysql.Config

	conn      int
	requests  uint64
	errors    uint64
	timeSpend int64
}

func (db *MonitoringWrapper) addRequest(err error, started time.Time) {
	db.requests++
	db.timeSpend += time.Since(started).Nanoseconds() / 1000000
	if err != nil {
		db.errors++
	}
}

func (db *MonitoringWrapper) getStatusRow() interfaces.StatusRow {
	return interfaces.StatusRow{
		Level: "",
		Data: []string{
			db.Conf.DBName,
			fmt.Sprintf("%d", db.conn),
			fmt.Sprintf("%d", db.requests),
			fmt.Sprintf("%d", db.errors),
			fmt.Sprintf("%0.2f", float64(db.timeSpend)/float64(db.requests)),
		},
	}
}

func NewMonitoringWrapper(db *sqlx.DB, dsn string) *MonitoringWrapper {
	conf, _ := mysql.ParseDSN(dsn)
	return &MonitoringWrapper{
		Db:   db,
		Conf: conf,
	}
}

func (this *MonitoringWrapper) GetHost() string {
	hostName := this.Conf.Addr
	i := strings.Index(hostName, ":")
	if i > 0 {
		hostName = hostName[:i]
	}
	return hostName
}

// NewDbAdapter creates new DB adapter.
func NewDbAdapter(dsns []string, tz string, debug bool) (*DbAdapter, error) {
	dbAdapter := &DbAdapter{
		debug: debug,
	}

	dbs := []*MonitoringWrapper{}

	var (
		errs []error
	)
	for _, dsn := range dsns {
		// ANSI_QUOTES — for compatibility with ANSI escaping with double quotes in queries
		dbConn, err := sqlx.Connect(driver, fmt.Sprintf("%s?parseTime=True&loc=%s&sql_mode=ANSI_QUOTES", dsn, url.QueryEscape(tz)))
		if err != nil {
			errs = append(errs, err)
		}
		monitoringWrapper := NewMonitoringWrapper(dbConn, dsn)
		dbs = append(dbs, monitoringWrapper)
		go func(monitoringWrapper *MonitoringWrapper) {
			for {
				if monitoringWrapper.Db == nil {
					return
				}
				time.Sleep(15 * time.Second)
				stats := monitoringWrapper.Db.Stats()
				mysqlmon.ConnectionNumber.WithLabelValues(monitoringWrapper.GetHost(), monitoringWrapper.Conf.DBName).Set(float64(stats.OpenConnections))
				monitoringWrapper.conn = stats.OpenConnections
			}
		}(monitoringWrapper)
	}

	dbAdapter.dbs = dbs

	if len(errs) > 0 {
		return dbAdapter, fmt.Errorf("DB connection errors: %v", errs)
	}

	return dbAdapter, nil
}

// Select rows from a DB.
func (this *DbAdapter) Select(dest interface{}, query string, args ...interface{}) error {
	started := time.Now()
	dbMonitoring := this.Slave()
	err := dbMonitoring.Db.Select(dest, query, args...)
	dbMonitoring.addRequest(err, started)
	mysqlmon.ResponseTime.WithLabelValues(dbMonitoring.GetHost(), dbMonitoring.Conf.DBName, metrics.IsError(err), "SELECT").Observe(metrics.SinceMs(started))
	return err

}

// Get one row from a DB.
func (this *DbAdapter) Get(dest interface{}, query string, args ...interface{}) error {
	started := time.Now()
	dbMonitoring := this.Slave()
	err := dbMonitoring.Db.Get(dest, query, args...)
	dbMonitoring.addRequest(err, started)
	return err
}

// Exec a query.
func (this *DbAdapter) Exec(query string, args map[string]interface{}) (sql.Result, error) {
	started := time.Now()
	dbMonitoring := this.Master()
	result, err := dbMonitoring.Db.NamedExec(query, args)
	dbMonitoring.addRequest(err, started)
	mysqlmon.ResponseTime.WithLabelValues(dbMonitoring.GetHost(), dbMonitoring.Conf.DBName, metrics.IsError(err), strings.SplitN(query, " ", 2)[0]).Observe(metrics.SinceMs(started))
	return result, err
}

// Master selects master server for queries.
func (this *DbAdapter) Master() *MonitoringWrapper {
	return this.dbs[0]
}

// Slave selects slave server for queries.
func (this *DbAdapter) Slave() *MonitoringWrapper {
	return this.dbs[this.slave(len(this.dbs))]
}

// Begin a transaction for a current session.
func (this *DbAdapter) Begin() (*sqlx.Tx, error) {
	return this.Master().Db.Beginx()
}

func (this *DbAdapter) slave(n int) int {
	if n <= 1 {
		return 0
	}
	return int(1 + (atomic.AddUint64(&this.count, 1) % uint64(n-1)))
}
