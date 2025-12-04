package modules

import (
	"github.com/dop251/goja"
)

type DelayedModule struct {
	poolSize int
	sem      chan struct{}
}

func NewDelayedModule(poolSize int) *DelayedModule {
	if poolSize <= 0 {
		poolSize = 10
	}
	return &DelayedModule{
		poolSize: poolSize,
		sem:      make(chan struct{}, poolSize),
	}
}

func (d *DelayedModule) Run(handler goja.Callable) {
	go func() {
		// Acquire semaphore
		d.sem <- struct{}{}
		defer func() {
			<-d.sem
			if r := recover(); r != nil {
				// Log panic but don't crash
			}
		}()

		handler(nil, nil)
	}()
}

// GetSchema implements JSSchemaProvider
func (d *DelayedModule) GetSchema() JSModuleSchema {
	return JSModuleSchema{
		Name:        "delayed",
		Description: "Run tasks asynchronously in background with worker pool limiting",
		Methods: []JSMethodSchema{
			{
				Name:        "run",
				Description: "Execute a function asynchronously in the background",
				Params:      []JSParamSchema{{Name: "handler", Type: "() => void", Description: "Function to execute"}},
			},
		},
	}
}

// GetDelayedSchema returns the delayed schema (static version)
func GetDelayedSchema() JSModuleSchema {
	return (&DelayedModule{}).GetSchema()
}
