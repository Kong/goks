package x509

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func pushErr(l *lua.LState, msg string) int {
	base := l.GetTop()
	l.Push(lua.LNil)
	l.Push(lua.LString(msg))
	return l.GetTop() - base
}

func validateCertificate(l *lua.LState) int {
	certPEM := l.CheckString(1)
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return pushErr(l, "failed to parse certificate")
	}

	_, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return pushErr(l, err.Error())
	}

	l.Push(lua.LTrue)
	return 1
}

func validateKey(l *lua.LState) int {
	keyPEM := l.CheckString(1)
	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return pushErr(l, "failed to parse key")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		if _, err := x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
			return pushErr(l, err.Error())
		}

	case "EC PRIVATE KEY":
		if _, err := x509.ParseECPrivateKey(block.Bytes); err != nil {
			return pushErr(l, err.Error())
		}

	case "PRIVATE KEY":
		if _, err := x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
			return pushErr(l, err.Error())
		}

	default:
		return pushErr(l, fmt.Sprintf("Unknown Key type %s", block.Type))
	}

	l.Push(lua.LTrue)
	return 1
}
