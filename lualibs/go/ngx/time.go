package ngx

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

func GetNgxTime(L *lua.LState) int {
	epoch := int32(time.Now().Unix())
	L.Push(lua.LNumber(epoch))
	return 1
}

func GetNgxNow(L *lua.LState) int {
	t := time.Now()
	seconds := t.Unix()
	miliseconds := t.UnixMilli() - seconds*1000
	result := float64(seconds) + float64(miliseconds)/float64(1000.0)
	L.Push(lua.LNumber(result))
	return 1
}

func NgxUpdateTime(L *lua.LState) int {
	return 0
}
