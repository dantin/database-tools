package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	ddl "github.com/dantin/database-tools/bulk_ddl"
	"github.com/juju/errors"
	"github.com/ngaut/log"
)

const (
	statChanSize  int = 1000
	queryChanSize int = 1000
)

func main() {
	cfg := ddl.NewConfig()
	err := cfg.Parse(os.Args[1:])
	switch errors.Cause(err) {
	case nil:
	case flag.ErrHelp:
		os.Exit(0)
	default:
		log.Errorf("parse cmd flags err %s\n", err)
		os.Exit(2)
	}
	log.SetLevelByString(cfg.LogLevel)

	log.Infof("config: %s", cfg)

	ddl.CreateConnPools(cfg)
	defer ddl.CloseConnPools(cfg)

	// Create a buffered channel to manage the task load.
	taskChan := make(map[string]chan *ddl.Job)
	cfg.StatChan = make(chan *ddl.Stat, statChanSize)

	for _, db := range cfg.ShardingDBs {
		taskChan[db] = make(chan *ddl.Job, queryChanSize)

		// Launch goroutines to handle the work.
		for i := 1; i <= cfg.WorkerCount; i++ {
			cfg.Wg.Add(1)
			go ddl.Worker(cfg, taskChan[db], i)
		}

		for i := 0; i < cfg.ShardingCount; i++ {
			t := &ddl.Job{
				Database: db,
				Table:    fmt.Sprintf("%s%d", cfg.ShardingTable, i),
			}
			taskChan[db] <- t
		}
	}

	cfg.WgStat.Add(1)
	startTs := time.Now()
	go ddl.StatWorker(cfg, &cfg.WgStat, startTs)

	for _, db := range cfg.ShardingDBs {
		close(taskChan[db])
	}
	cfg.Wg.Wait()
	close(cfg.StatChan)
	cfg.WgStat.Wait()

	log.Info("Done!")
}
