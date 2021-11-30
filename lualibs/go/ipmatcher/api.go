package ipmatcher

import (
	"net"

	lua "github.com/yuin/gopher-lua"
)

func ParseIPv4(L *lua.LState) int {
	input := L.CheckString(1)
	ip := net.ParseIP(input)
	if ip == nil {
		L.Push(lua.LBool(false))
		return 1
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		L.Push(lua.LBool(false))
		return 1
	}
	L.Push(lua.LBool(true))
	return 1
}

func ParseIPv6(L *lua.LState) int {
	input := L.CheckString(1)
	ip := net.ParseIP(input)
	if ip == nil {
		L.Push(lua.LBool(false))
		return 1
	}
	// TODO(hbagdi): figure out a better way to ensure that this IP is a v6
	ipv4 := ip.To4()
	if ipv4 != nil {
		L.Push(lua.LBool(false))
		return 1
	}
	L.Push(lua.LBool(true))
	return 1
}
