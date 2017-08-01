package bulk_ddl

import (
	"flag"
	"fmt"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/dantin/database-tools/utils"
	"github.com/juju/errors"
	"github.com/ngaut/pool"
)

func NewConfig() *Config {
	cfg := &Config{}
	cfg.FlagSet = flag.NewFlagSet("bulk_ddl", flag.ContinueOnError)
	fs := cfg.FlagSet

	fs.StringVar(&cfg.configFile, "config", "", "Config File")

	fs.IntVar(&cfg.WorkerCount, "c", 40, "parallel worker count per data source")
	fs.IntVar(&cfg.PoolSize, "pool", 40, "db connection pool size per data source")
	fs.IntVar(&cfg.reportInterval, "report-interval", 1, "report status interval")

	fs.BoolVar(&cfg.rollback, "R", false, "rollback flag")
	fs.BoolVar(&cfg.printVersion, "V", false, "print version and exit")

	fs.StringVar(&cfg.LogLevel, "L", "info", "log level: debug, info, warn, error, fatal")

	return cfg
}

// Config is the configuration.
type Config struct {
	*flag.FlagSet `json:"-"`

	DBs []*DBConfig `toml:"dbs" json:"dbs"`

	ShardingDB    string `toml:"sharding-db" json:"sharding-db"`
	ShardingTable string `toml:"sharding-table" json:"sharding-table"`
	ShardingCount int    `toml:"sharding-count" json:"sharding-count"`

	DoDDL   string `toml:"do-ddl" json:"do-ddl"`
	Token   string `toml:"token" json:"token"`
	UndoDDL string `toml:"undo-ddl" json:"undo-ddl"`

	WorkerCount int `toml:"worker-count" json:"worker-count"`
	PoolSize    int `toml:"pool-size" json:"pool-size"`

	reportInterval int `toml:"report-interval" json:"report-interval"`

	LogLevel string `toml:"log-level" json:"log-level"`

	ShardingDBs []string

	configFile   string
	printVersion bool
	rollback     bool

	connPools map[string]*pool.Cache

	Wg     sync.WaitGroup
	WgStat sync.WaitGroup

	StatChan chan *Stat
}

func (c Config) String() string {
	dbs := make([]string, len(c.DBs))
	for i, db := range c.DBs {
		dbs[i] = fmt.Sprintf("%+v", *db)
	}

	dbsStr := fmt.Sprintf("[%s]", strings.Join(dbs, ";"))

	return fmt.Sprintf("worker-count:%d sharding-db:%s sharding-table:%s sharding-count:%d "+
		"do-ddl:%s undo-ddl:%s token:%s log-level: %s report-interval: %d dbs:%s",
		c.WorkerCount, c.ShardingDB, c.ShardingTable, c.ShardingCount, c.DoDDL, c.UndoDDL, c.Token, c.LogLevel, c.reportInterval, dbsStr)
}

// Parse parse flag definitions from the argument list.
func (c *Config) Parse(arguments []string) error {
	// Parse first to get config file.
	err := c.FlagSet.Parse(arguments)
	if err != nil {
		return errors.Trace(err)
	}

	if c.printVersion {
		fmt.Printf(utils.GetRawInfo("bulk_ddl"))
		return flag.ErrHelp
	}

	// Load config file if specified.
	if c.configFile != "" {
		err = c.configFromFile(c.configFile)
		if err != nil {
			return errors.Trace(err)
		}
	}

	err = c.FlagSet.Parse(arguments)
	if err != nil {
		return errors.Trace(err)
	}

	if len(c.FlagSet.Args()) != 0 {
		return errors.Errorf("'%s' is an invalid flag", c.FlagSet.Arg(0))
	}

	c.ShardingDBs = make([]string, 0)
	for _, db := range strings.Split(c.ShardingDB, ",") {
		c.ShardingDBs = append(c.ShardingDBs, strings.Trim(db, " "))
	}

	c.Token = strings.ToUpper(c.Token)

	return nil
}

// DBConfig is the DB configuration.
type DBConfig struct {
	Host     string `toml:"host" json:"host"`
	User     string `toml:"user" json:"user"`
	Password string `toml:"password" json:"password"`
	Name     string `toml:"name" json:"name"`
	Port     int    `toml:"port" json:"port"`
}

func (c *DBConfig) String() string {
	if c == nil {
		return "<nil>"
	}

	return fmt.Sprintf("DBConfig(%+v)", *c)
}

// configFromFile loads config from file.
func (c *Config) configFromFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	return errors.Trace(err)
}
