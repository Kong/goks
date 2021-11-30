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

func ValidateUUID(L *lua.LState) int {
	input := L.CheckString(1)
	// googleUUID.Parse accepts other formats that shouldn't be allowed in this
	// context.
	// the following string len check ensures this
	if len(input) != 36 {
		L.Push(lua.LBool(false))
		return 1
	}
	_, err := googleUUID.Parse(input)
	if err != nil {
		L.Push(lua.LBool(false))
		return 1
	}
	L.Push(lua.LBool(true))
	return 1
}
