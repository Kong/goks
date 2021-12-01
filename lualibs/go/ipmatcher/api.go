package ipmatcher

import (
	"net"

	lua "github.com/yuin/gopher-lua"
)

func ParseIPv4(l *lua.LState) int {
	input := l.CheckString(1)
	ip := net.ParseIP(input)
	if ip == nil {
		l.Push(lua.LBool(false))
		return 1
	}
	if ipv4 := ip.To4(); ipv4 == nil {
		l.Push(lua.LBool(false))
		return 1
	}
	l.Push(lua.LBool(true))
	return 1
}

func ParseIPv6(l *lua.LState) int {
	input := l.CheckString(1)
	ip := net.ParseIP(input)
	if ip == nil {
		l.Push(lua.LBool(false))
		return 1
	}
	// TODO(hbagdi): figure out a better way to ensure that this IP is a v6
	if ipv4 := ip.To4(); ipv4 != nil {
		l.Push(lua.LBool(false))
		return 1
	}
	l.Push(lua.LBool(true))
	return 1
}
