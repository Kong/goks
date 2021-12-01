package re2

import (
	lua "github.com/yuin/gopher-lua"
)

func Loader(L *lua.LState) int {
	t := L.NewTable()
	L.SetFuncs(t, api)
	L.Push(t)
	return 1
}

var api = map[string]lua.LGFunction{
	"find":    ParseIPv4,
	"compile": ParseIPv6,
}
