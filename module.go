package data

import "github.com/farseer-go/fs/modules"

type Module struct {
}

func (module Module) DependsModule() []modules.FarseerModule {
	return nil
}

func (module Module) PreInitialize() {
}

func (module Module) Initialize() {
}

func (module Module) PostInitialize() {
	checkConfig()
}

func (module Module) Shutdown() {
}
