package ngx

import (
	lua "github.com/yuin/gopher-lua"
)

func LoadNgx(L *lua.LState) {
	ngx := L.NewTable()
	// register other stuff
	L.SetFuncs(ngx, api)
	L.SetField(ngx, "null", lua.LString("null"))
	L.SetGlobal("ngx", ngx)
}

var api = map[string]lua.LGFunction{
	"time":          GetNgxTime,
	"now":           GetNgxNow,
	"update_time":   NgxUpdateTime,
	"encode_base64": EncodeBase64,
}
