package ngx

import (
	lua "github.com/yuin/gopher-lua"
)

func LoadNgx(l *lua.LState) {
	ngx := l.NewTable()
	// register other stuff
	l.SetFuncs(ngx, api)
	l.SetField(ngx, "null", lua.LString("null"))
	l.SetGlobal("ngx", ngx)
}

var api = map[string]lua.LGFunction{
	"time":          GetNgxTime,
	"now":           GetNgxNow,
	"update_time":   UpdateTime,
	"encode_base64": EncodeBase64,
}
