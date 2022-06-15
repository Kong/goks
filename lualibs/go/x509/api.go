package x509

import (
	"crypto/x509"
	"encoding/pem"

	lua "github.com/yuin/gopher-lua"
)

func pushErr(l *lua.LState, msg string) int {
	l.Push(lua.LNil)
	l.Push(lua.LString(msg))
	return 2
}

func validateCertificate(l *lua.LState) int {
	certPEM := l.CheckString(1)
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return pushErr(l, "failed to parse certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return pushErr(l, err.Error())
	}

	_, err = cert.Verify(x509.VerifyOptions{})
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

	_, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return pushErr(l, err.Error())
	}

	l.Push(lua.LTrue)
	return 1
}
