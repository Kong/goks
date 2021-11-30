package ngx

import (
	lua "github.com/yuin/gopher-lua"
)

func LoadNgx(L *lua.LState) {
	ngx := L.NewTable()
	// register other stuff
	L.SetField(ngx, "null", lua.LString("null"))
	L.SetGlobal("ngx", ngx)
}
