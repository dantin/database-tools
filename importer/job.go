package importer

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/juju/errors"
	"github.com/ngaut/log"
)

// Process populate data task
func DoProcess(table *Table, dbs []*sql.DB, jobCount int, workerCount int, batch int) {
	// channel with buffer, like message queue
	jobChan := make(chan struct{}, 16*workerCount)
	doneChan := make(chan struct{}, workerCount)

	start := time.Now()
	go addJobs(jobCount, jobChan)

	for i := 0; i < workerCount; i++ {
		go doJob(table, dbs[i], batch, jobChan, doneChan)
	}

	doWait(doneChan, start, jobCount, workerCount)
}

// Add job to channel, works like a dispatcher
func addJobs(jobCount int, jobChan chan struct{}) {
	for i := 0; i < jobCount; i++ {
		jobChan <- struct{}{}
	}

	// close the job channel(not kill it), like EOF
	close(jobChan)
}

// Do job task
func doJob(table *Table, db *sql.DB, batch int, jobChan chan struct{}, doneChan chan struct{}) {
	count := 0

	for range jobChan {
		count++
		if count == batch {
			doInsert(table, db, count)
			count = 0
		}
	}

	if count > 0 {
		doInsert(table, db, count)
		count = 0
	}

	// send notification
	doneChan <- struct{}{}
}

func doInsert(table *Table, db *sql.DB, count int) {
	// generate row SQL statement
	rows, err := genRowsData(table, count)
	if err != nil {
		log.Fatalf(errors.ErrorStack(err))
	}

	txn, err := db.Begin()
	if err != nil {
		log.Fatalf(errors.ErrorStack(err))
	}

	for _, row := range rows {
		_, err = txn.Exec(row)
		if err != nil {
			log.Fatalf(errors.ErrorStack(err))
		}
	}

	err = txn.Commit()
	if err != nil {
		log.Fatalf(errors.ErrorStack(err))
	}
}

func doWait(doneChan chan struct{}, start time.Time, jobCount int, workerCount int) {
	for i := 0; i < workerCount; i++ {
		<-doneChan
	}

	close(doneChan)

	now := time.Now()
	seconds := now.Unix() - start.Unix()

	tps := int64(-1)
	if seconds > 0 {
		tps = int64(jobCount) / seconds
	}

	fmt.Printf("[importer]total %d cases, cost %d seconds, tps %d, start %s, now %s\n",
		jobCount, seconds, tps, start, now)
}

func genRowsData(table *Table, count int) ([]string, error) {
	rows := make([]string, 0, count)
	for i := 0; i < count; i++ {
		data, err := GenRowData(table)
		if err != nil {
			return nil, errors.Trace(err)
		}
		rows = append(rows, data)
	}

	return rows, nil
}
