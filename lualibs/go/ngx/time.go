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
	now := Now()
	l.Push(lua.LNumber(now))
	return 1
}

func Now() float64 {
	t := time.Now()
	seconds := t.Unix()
	var secondMultiplier int64 = 1000
	msDivisor := 1000.0
	miliseconds := t.UnixMilli() - seconds*secondMultiplier
	return float64(seconds) + float64(miliseconds)/msDivisor
}

func UpdateTime(_ *lua.LState) int {
	return 0
}
