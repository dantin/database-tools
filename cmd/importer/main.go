package main

import (
	"flag"
	"os"

	"github.com/dantin/database-tools/importer"
	"github.com/juju/errors"
	"github.com/ngaut/log"
)

func main() {
	cfg := importer.NewConfig()
	err := cfg.Parse(os.Args[1:])
	switch errors.Cause(err) {
	case nil:
	case flag.ErrHelp:
		os.Exit(0)
	default:
		log.Errorf("parse cmd flag err %s\n", err)
		os.Exit(2)
	}

	table := importer.NewTable()
	err = importer.ParseTableSQL(table, cfg.TableSQL)
	if err != nil {
		log.Fatal(err)
	}

	err = importer.ParseIndexSQL(table, cfg.IndexSQL)
	if err != nil {
		log.Fatal(err)
	}

	dbs, err := importer.CreateDBs(cfg.DBCfg, cfg.WorkerCount)
	if err != nil {
		log.Fatal(err)
	}
	defer importer.CloseDBs(dbs)

	err = importer.ExecSQL(dbs[0], cfg.TableSQL)
	if err != nil {
		log.Fatal(err)
	}

	err = importer.ExecSQL(dbs[0], cfg.IndexSQL)
	if err != nil {
		log.Fatal(err)
	}

	importer.DoProcess(table, dbs, cfg.JobCount, cfg.WorkerCount, cfg.Batch)
}
