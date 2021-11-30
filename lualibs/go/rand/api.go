package rand

import (
	cryptoRand "crypto/rand"

	lua "github.com/yuin/gopher-lua"
)

// MD5 lua crypto.md5(string) return string
func GetRandBytes(L *lua.LState) int {
	size := L.CheckInt(1)
	buffer := make([]byte, size)
	_, err := cryptoRand.Read(buffer)
	if err != nil {
		panic(err)
	}
	L.Push(lua.LString(string(buffer)))
	return 1
}
