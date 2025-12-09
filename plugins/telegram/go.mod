module telegram

go 1.24.0

toolchain go1.24.9

require (
	github.com/dop251/goja v0.0.0-20241024094426-79f3a7efcdbd
	github.com/go-telegram/bot v1.11.1
	github.com/spf13/cast v1.7.0
	m3m v0.0.0
)

require (
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/google/pprof v0.0.0-20230207041349-798e818bf904 // indirect
	golang.org/x/text v0.31.0 // indirect
)

replace m3m => ../..
