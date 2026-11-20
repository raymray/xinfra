package xdb

import (
	"sort"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// dialectors 是 driver 名到 Dialector 构造函数的注册表，与 database/sql 的驱动注册同理。
// 新增国产库时在 init 中调用 RegisterDialector，业务代码无需改动。
var (
	dialectMu  sync.RWMutex
	dialectors = map[string]func(dsn string) gorm.Dialector{
		DriverMySQL:    mysql.Open,
		DriverPostgres: postgres.Open,
	}
)

func RegisterDialector(driver string, open func(dsn string) gorm.Dialector) {
	dialectMu.Lock()
	defer dialectMu.Unlock()
	dialectors[driver] = open
}

func hasDialector(driver string) bool {
	dialectMu.RLock()
	defer dialectMu.RUnlock()
	_, ok := dialectors[driver]
	return ok
}

func dialector(driver, dsn string) (gorm.Dialector, bool) {
	dialectMu.RLock()
	defer dialectMu.RUnlock()
	open, ok := dialectors[driver]
	if !ok {
		return nil, false
	}
	return open(dsn), true
}

func registeredDrivers() []string {
	dialectMu.RLock()
	defer dialectMu.RUnlock()
	names := make([]string, 0, len(dialectors))
	for k := range dialectors {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
