package rand

import (
	cryptoRand "crypto/rand"

	lua "github.com/yuin/gopher-lua"
)

func GetRandBytes(l *lua.LState) int {
	size := l.CheckInt(1)
	buffer := make([]byte, size)
	if _, err := cryptoRand.Read(buffer); err != nil {
		panic(err)
	}
	l.Push(lua.LString(string(buffer)))
	return 1
}
