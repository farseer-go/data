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
}
func (module Module) PostInitialize() {
	nodes := configure.GetSubNodes("Database")
	for key, val := range nodes {
		configString := val.(string)
		if configString == "" {
			panic("[farseer.yaml]Database." + key + "，没有正确配置")
		}
		config := configure.ParseString[dbConfig](configString)
		if config.ConnectionString == "" {
			panic("[farseer.yaml]Database." + key + ".ConnectionString，没有正确配置")
		}
		if config.DataType == "" {
			panic("[farseer.yaml]Database." + key + ".DataType，没有正确配置")
		}

		// 注册健康检查
		container.RegisterInstance[core.IHealthCheck](&healthCheck{name: key}, "db_"+key)
	}
}
