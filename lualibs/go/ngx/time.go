package ngx

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

func GetNgxTime(l *lua.LState) int {
	epoch := int32(time.Now().Unix())
	l.Push(lua.LNumber(epoch))
	return 1
}

func GetNgxNow(l *lua.LState) int {
	t := time.Now()
	seconds := t.Unix()
	var secondMultiplier int64 = 1000
	msDivisor := 1000.0
	miliseconds := t.UnixMilli() - seconds*secondMultiplier
	result := float64(seconds) + float64(miliseconds)/msDivisor
	l.Push(lua.LNumber(result))
	return 1
}

func UpdateTime(_ *lua.LState) int {
	return 0
}
