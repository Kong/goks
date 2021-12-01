package ipmatcher

import (
	lua "github.com/yuin/gopher-lua"
)

func Loader(l *lua.LState) int {
	t := l.NewTable()
	l.SetFuncs(t, api)
	l.Push(t)
	return 1
}

var api = map[string]lua.LGFunction{
	"parse_ipv4": ParseIPv4,
	"parse_ipv6": ParseIPv6,
}
