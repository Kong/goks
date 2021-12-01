package ngx

import (
	"encoding/base64"

	lua "github.com/yuin/gopher-lua"
)

func EncodeBase64(l *lua.LState) int {
	input := l.CheckString(1)
	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	l.Push(lua.LString(encoded))
	return 1
}
