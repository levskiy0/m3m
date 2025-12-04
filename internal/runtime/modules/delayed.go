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
