package bulk_ddl

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dantin/database-tools/utils"
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/ngaut/pool"
)

// Structure for job.
type Job struct {
	Database string
	Table    string
}

// Structure for stat result.
type Stat struct {
	spend time.Duration
	succ  bool
}

func (j Job) String() string {
	return fmt.Sprintf("database: %s table: %s", j.Database, j.Table)
}

func Worker(cfg *Config, tasks chan *Job, worker int) {
	defer cfg.Wg.Done()

	for {
		// Wait for work to be assigned.
		task, ok := <-tasks
		if !ok {
			// This means the channel is empty and closed.
			log.Debugf("Worker: %d : Shutting Down", worker)
			return
		}

		// Display we are starting the work.
		log.Debugf("Worker: %d : Started %s", worker, task)

		columns, err := getTableColumns(cfg, cfg.connPools[task.Database], task.Database, task.Table)

		if err != nil {
			log.Warn(err)
		} else {
			var sqlStmt string

			if !cfg.rollback && !utils.StringInSlice(cfg.Token, columns) {
				sqlStmt = fmt.Sprintf(cfg.DoDDL, task.Database, task.Table)
				log.Debugf("Database: %s, Table: %s: do ddl", task.Database, task.Table)
				exec(cfg, cfg.connPools[task.Database], sqlStmt)
			} else if cfg.rollback && utils.StringInSlice(cfg.Token, columns) {
				sqlStmt = fmt.Sprintf(cfg.UndoDDL, task.Database, task.Table)
				log.Debugf("Database: %s, Table: %s: undo ddl", task.Database, task.Table)
				exec(cfg, cfg.connPools[task.Database], sqlStmt)
			} else {
				log.Debugf("Database: %s, Table: %s, columns: %v", task.Database, task.Table, columns)
			}
		}

		// Display we finished the work.
		log.Debugf("Worker: %d : Completed %v", worker, task)
	}
}

func exec(cfg *Config, connPool *pool.Cache, sqlStmt string) (*sql.Rows, error) {
	stmt := strings.ToLower(sqlStmt)
	isQuery := strings.HasPrefix(stmt, "select")

	db, ok := connPool.Get().(*sql.DB)
	if !ok {
		log.Fatal("The type of db got from pool is wrong")
	}
	defer connPool.Put(db)

	// Get time
	startTs := time.Now()
	rows, err := runQuery(db, stmt, isQuery)
	if err != nil {
		log.Warnf("Exec sql [%s]: %s", sqlStmt, err)
		cfg.StatChan <- &Stat{}
		return nil, err
	}

	cfg.StatChan <- &Stat{spend: time.Now().Sub(startTs), succ: true}
	return rows, nil
}

func runQuery(db *sql.DB, sqlStmt string, isQuery bool) (*sql.Rows, error) {
	if isQuery {
		rows, err := db.Query(sqlStmt)
		if err != nil {
			log.Fatalf(errors.ErrorStack(err))
			return nil, err
		}
		return rows, nil
	}

	_, err := db.Exec(sqlStmt)
	return nil, err
}

func getTableColumns(cfg *Config, connPool *pool.Cache, database string, table string) ([]string, error) {
	columns := make([]string, 0)

	sql := fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_schema='%s' AND table_name='%s'", database, table)
	rows, err := exec(cfg, connPool, sql)
	if err != nil {
		return columns, errors.Trace(err)
	}
	defer rows.Close()

	for rows.Next() {
		var column string
		err = rows.Scan(&column)
		column = strings.ToUpper(column)
		if err != nil {
			return columns, errors.Trace(err)
		}

		columns = append(columns, column)
	}

	if rows.Err() != nil {
		return columns, errors.Trace(rows.Err())
	}

	return columns, nil
}

func StatWorker(cfg *Config, wg *sync.WaitGroup, startTs time.Time) {
	defer wg.Done()
	var (
		total       int64
		succ        int64
		spend       time.Duration
		tempStartTs = startTs
		tempTotal   int64
		tempSpend   time.Duration
		tempSucc    int64
	)

	for {
		tempExecTime := time.Now().Sub(tempStartTs)
		if cfg.reportInterval != 0 && tempExecTime.Seconds() >= float64(cfg.reportInterval) {
			log.Infof("Query: %d, Succ: %d, Faild: %d, Time: %v, Avg response time: %.04fms, QPS: %.02f : \n", tempTotal, tempSucc, tempTotal-tempSucc, tempExecTime, (tempSpend.Seconds()*1000)/float64(tempTotal), float64(tempTotal)/tempExecTime.Seconds())
			tempStartTs = time.Now()
			tempTotal = 0
			tempSpend = 0
			tempSucc = 0
		}
		s, ok := <-cfg.StatChan
		if !ok {
			break
		}
		total++
		tempTotal++
		if s.succ {
			succ++
			tempSucc++
		}
		spend += s.spend
		tempSpend += s.spend
	}
	execTime := time.Now().Sub(startTs)
	log.Info("\n*************************final result***************************\n")
	log.Infof("Total Query: %d, Succ: %d, Faild: %d, Time: %v, Avg response time: %.04fms, QPS: %.02f : \n", total, succ, total-succ, execTime, (spend.Seconds()*1000)/float64(total), float64(total)/execTime.Seconds())
}
