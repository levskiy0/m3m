package modules

type EnvModule struct {
	vars map[string]interface{}
}

func NewEnvModule(vars map[string]interface{}) *EnvModule {
	if vars == nil {
		vars = make(map[string]interface{})
	}
	return &EnvModule{vars: vars}
}

func (e *EnvModule) Get(key string) interface{} {
	return e.vars[key]
}
