package x509

import (
	lua "github.com/yuin/gopher-lua"
)

func Loader(l *lua.LState) int {
	t := l.NewTable()
	l.SetFuncs(t, api)
	l.Push(t)
	return 1
}

var api = map[string]lua.LGFunction{
	"validate_certificate": validateCertificate,
	"validate_key":         validateKey,
}
