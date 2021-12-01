package uuid

import (
	googleUUID "github.com/google/uuid"
	lua "github.com/yuin/gopher-lua"
)

func GetUUID(l *lua.LState) int {
	uuid := googleUUID.NewString()
	l.Push(lua.LString(uuid))
	return 1
}

const uuidLen = 36

func ValidateUUID(l *lua.LState) int {
	input := l.CheckString(1)
	// googleUUID.Parse accepts other formats that shouldn't be allowed in this
	// context.
	// the following string len check ensures this
	if len(input) != uuidLen {
		l.Push(lua.LBool(false))
		return 1
	}
	if _, err := googleUUID.Parse(input); err != nil {
		l.Push(lua.LBool(false))
		return 1
	}
	l.Push(lua.LBool(true))
	return 1
}
