package ipmatcher

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
	"parse_ipv4": ParseIPv4,
	"parse_ipv6": ParseIPv6,
}
