package importer

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/juju/errors"
	"github.com/ngaut/log"
)

// Create DB connections
func CreateDBs(cfg DBConfig, count int) ([]*sql.DB, error) {
	dbs := make([]*sql.DB, 0, count)
	for i := 0; i < count; i++ {
		db, err := createDB(cfg)
		if err != nil {
			return nil, errors.Trace(err)
		}

		dbs = append(dbs, db)
	}

	return dbs, nil
}

func createDB(cfg DBConfig) (*sql.DB, error) {
	// prepare data source name
	dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return db, nil
}

// Close DB connections
func CloseDBs(dbs []*sql.DB) {
	for _, db := range dbs {
		err := closeDB(db)
		if err != nil {
			log.Errorf("close db failed - %v", err)
		}
	}
}

func closeDB(db *sql.DB) error {
	return errors.Trace(db.Close())
}

// Execute SQL
func ExecSQL(db *sql.DB, sql string) error {
	if len(sql) == 0 {
		return nil
	}

	_, err := db.Exec(sql)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}
