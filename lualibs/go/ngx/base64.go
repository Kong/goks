package ngx

import (
	"encoding/base64"

	lua "github.com/yuin/gopher-lua"
)

func EncodeBase64(L *lua.LState) int {
	input := L.CheckString(1)
	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	L.Push(lua.LString(encoded))
	return 1
}
