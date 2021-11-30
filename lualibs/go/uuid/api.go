package uuid

import (
	googleUUID "github.com/google/uuid"
	lua "github.com/yuin/gopher-lua"
)

func GetUUID(L *lua.LState) int {
	uuid := googleUUID.NewString()
	L.Push(lua.LString(uuid))
	return 1
}
