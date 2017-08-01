package bulk_ddl

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ngaut/log"
	"github.com/ngaut/pool"
)

func createConnPool(cfg DBConfig, poolSize int) *pool.Cache {
	connPool := pool.NewCache(fmt.Sprintf("%s-pool", cfg.Name), poolSize, func() interface{} {
		dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
		db, err := sql.Open("mysql", dbDSN)
		if err != nil {
			log.Fatal(err)
		}
		var tag bool
		for i := 0; i < 10; i++ {
			err = db.Ping()
			if err == nil {
				tag = true
				break
			}
			log.Warnf("%s db.ping failed %v", cfg.Name, err)
		}
		if tag == false {
			log.Fatal("%s break donw", cfg.Name)
		}

		return db
	})

	return connPool
}

func closeConnPool(poolSize int, connPool *pool.Cache) {
	for i := 0; i < poolSize; i++ {
		db, ok := connPool.Get().(*sql.DB)
		if !ok {
			log.Fatal("cast error!")
		}
		db.Close()
	}
}

func CreateConnPools(cfg *Config) {
	dbs := make(map[string]*pool.Cache)

	for _, key := range cfg.ShardingDBs {
		for _, dbCfg := range cfg.DBs {
			if dbCfg.Name == key {
				connPool := createConnPool(*dbCfg, cfg.PoolSize)
				dbs[key] = connPool
			}
		}
	}

	cfg.connPools = dbs
}

func CloseConnPools(cfg *Config) {
	log.Info("close connection pool")

	for _, connPool := range cfg.connPools {
		closeConnPool(cfg.PoolSize, connPool)
	}
}
