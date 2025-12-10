package jsutil

import (
	"github.com/dop251/goja"
)

// ThrowJSError throws a JavaScript Error that can be caught with try-catch.
// This should be used instead of panic(string) when you want the error
// to be catchable in JavaScript code.
func ThrowJSError(vm *goja.Runtime, msg string) {
	if vm != nil {
		errConstructor := vm.Get("Error")
		if errConstructor != nil {
			if errFunc, ok := goja.AssertFunction(errConstructor); ok {
				errObj, _ := errFunc(goja.Undefined(), vm.ToValue(msg))
				panic(errObj)
			}
		}
	}
	panic(msg)
}
