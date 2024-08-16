package pdsql

import (
	"fmt"
	"log"

	"github.com/wenerme/coredns-pdsql/pdnsmodel"

	// for now we will just add all the drivers and switch on them

	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"gorm.io/gorm"
)

func init() {
	caddy.RegisterPlugin("pdsql", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	backend := PowerDNSGenericSQLBackend{}
	c.Next()
	if !c.NextArg() {
		return plugin.Error("pdsql", c.ArgErr())
	}
	dialect := c.Val()

	if !c.NextArg() {
		return plugin.Error("pdsql", c.ArgErr())
	}
	arg := c.Val()
	fmt.Println("dialect: ", dialect)
	fmt.Println("arg: ", arg)

	var d gorm.Dialector

	switch dialect {
	case "mysql":
		d = mysql.Open(arg)
	case "postgres":
		d = postgres.Open(arg)
	case "clickhouse ":
		d = clickhouse.Open(arg)
	case "sqlite":
		d = sqlite.Open(arg)
	case "sqlserver":
		d = sqlserver.Open(arg)

	}

	db, err := gorm.Open(d, &gorm.Config{})
	if err != nil {
		return err
	}
	backend.DB = db

	for c.NextBlock() {
		x := c.Val()
		switch x {
		case "debug":
			args := c.RemainingArgs()
			for _, v := range args {
				switch v {
				case "db":
					backend.DB = backend.DB.Debug()
				}
			}
			backend.Debug = true
			log.Println(Name, "enable log", args)
		case "auto-migrate":
			// currently only use records table
			if err := backend.AutoMigrate(); err != nil {
				return err
			}
		default:
			return plugin.Error("pdsql", c.Errf("unexpected '%v' command", x))
		}
	}

	if c.NextArg() {
		return plugin.Error("pdsql", c.ArgErr())
	}

	dnsserver.
		GetConfig(c).
		AddPlugin(func(next plugin.Handler) plugin.Handler {
			backend.Next = next
			return backend
		})

	return nil
}

func (pdb PowerDNSGenericSQLBackend) AutoMigrate() error {
	return pdb.DB.AutoMigrate(&pdnsmodel.Record{}, &pdnsmodel.Domain{})
}
