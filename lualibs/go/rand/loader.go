package rand

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
	"get_rand_bytes": GetRandBytes,
}
