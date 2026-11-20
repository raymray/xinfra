package xdb

import (
	"context"
	"time"

	"github.com/raymray/xinfra/xerror"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Client interface {
	DB() *gorm.DB
	// ToSQL 返回 DryRun 会话，只生成 SQL 不执行
	ToSQL() *gorm.DB
	Ping() error
	Close() error
}

type client struct {
	db          *gorm.DB
	pingTimeout time.Duration
}

// newClient 按已填充默认值的配置建连并 Ping。
func newClient(cfg Config, logger gormlogger.Interface) (*client, error) {
	dial, ok := dialector(cfg.Driver, cfg.DSN)
	if !ok {
		return nil, xerror.Newf("unsupported db driver: %s (registered: %v)", cfg.Driver, registeredDrivers())
	}
	db, err := gorm.Open(dial, &gorm.Config{Logger: logger})
	if err != nil {
		return nil, xerror.Wrapf(err, "open db")
	}
	sdb, err := db.DB()
	if err != nil {
		return nil, xerror.Wrapf(err, "underlying sql.DB")
	}
	sdb.SetMaxOpenConns(cfg.MaxOpenConns)
	sdb.SetMaxIdleConns(cfg.MaxIdleConns)
	sdb.SetConnMaxLifetime(time.Duration(cfg.MaxLifeTime) * time.Second)
	sdb.SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second)

	c := &client{db: db, pingTimeout: time.Duration(cfg.DialTimeout) * time.Second}
	if err := c.Ping(); err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

func (c *client) DB() *gorm.DB { return c.db }

func (c *client) ToSQL() *gorm.DB {
	return c.db.Session(&gorm.Session{DryRun: true, SkipDefaultTransaction: true})
}

func (c *client) Ping() error {
	sdb, err := c.db.DB()
	if err != nil {
		return xerror.Wrapf(err, "underlying sql.DB")
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.pingTimeout)
	defer cancel()
	if err := sdb.PingContext(ctx); err != nil {
		return xerror.Wrapf(err, "ping db")
	}
	return nil
}

func (c *client) Close() error {
	sdb, err := c.db.DB()
	if err != nil {
		return xerror.Wrapf(err, "underlying sql.DB")
	}
	return sdb.Close()
}
