package sqlserver

import (
	"github.com/farseer-go/data"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/modules"
)

type Module struct {
}

func (module Module) DependsModule() []modules.FarseerModule {
	return []modules.FarseerModule{data.Module{}}
}

func (module Module) Initialize() {
	container.Register(func() data.IDataDriver {
		return &dataDriver{}
	}, "sqlserver")
}
