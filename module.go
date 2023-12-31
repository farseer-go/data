package data

import (
	"github.com/farseer-go/data/driver/mysql"
	"github.com/farseer-go/fs/configure"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/modules"
	"gorm.io/gorm"
)

type Module struct {
}

func (module Module) DependsModule() []modules.FarseerModule {
	return nil
}

func (module Module) PreInitialize() {
	databaseConn = make(map[string]*gorm.DB)
}

func (module Module) Initialize() {
	nodes := configure.GetSubNodes("Database")
	for key, val := range nodes {
		configString := val.(string)
		if configString == "" {
			panic("[farseer.yaml]Database." + key + "，配置不正确")
		}
		// 注册内部上下文
		RegisterInternalContext(key, configString)
	}

	// 注册mysql驱动
	container.Register(func() IDataDriver {
		return &mysql.DataDriver{}
	}, "mysql")
}
