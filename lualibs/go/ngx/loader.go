package ngx

import (
	lua "github.com/yuin/gopher-lua"
)

func NullString(l *lua.LState) int {
	l.Push(lua.LString("null"))
	return 1
}

func LoadNgx(l *lua.LState) {
	ngx := l.NewTable()
	l.SetFuncs(ngx, api)

	// register ngx.null
	d := l.NewUserData()
	d.Metatable = l.NewTable()
	mt := l.NewTable()
	l.SetFuncs(mt, map[string]lua.LGFunction{"__tostring": NullString})
	l.SetField(ngx, "null", d)
	l.SetGlobal("ngx", ngx)
}

var api = map[string]lua.LGFunction{
	"time":          GetNgxTime,
	"now":           GetNgxNow,
	"update_time":   UpdateTime,
	"encode_base64": EncodeBase64,
}
