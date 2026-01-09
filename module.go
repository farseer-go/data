package data

import (
	"github.com/farseer-go/fs/configure"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/core"
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
	// 注册包级别的连接检查器（默认实现）
	container.Register(func() core.IConnectionChecker { return &connectionChecker{} }, "data")
}

func (module Module) Initialize() {
	// 注册mysql驱动
	container.Register(func() IDataDriver { return &DataDriver{} }, "mysql")

	nodes := configure.GetSubNodes("Database")
	for key, val := range nodes {
		configString := val.(string)
		if configString == "" {
			panic("[farseer.yaml]Database." + key + "，配置不正确")
		}
		// 注册内部上下文
		RegisterInternalContext(key, configString)
	}
}
