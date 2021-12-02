package vm

import (
	"fmt"
	"io"
	"sync"

	"github.com/kong/goks"
	"github.com/kong/goks/internal/fs"
	"github.com/kong/goks/lualibs/go/ipmatcher"
	"github.com/kong/goks/lualibs/go/ngx"
	"github.com/kong/goks/lualibs/go/rand"
	"github.com/kong/goks/lualibs/go/uuid"
	json "github.com/layeh/gopher-json"
	"github.com/yuin/gluare"
	lua "github.com/yuin/gopher-lua"
)

type VM struct {
	l  *lua.LState
	mu sync.Mutex
}

func New() (*VM, error) {
	LuaLDir := "lua-tree/share/lua/5.1"
	lua.LuaPathDefault = "/?.lua;" +
		LuaLDir + "/?.lua;" +
		LuaLDir + "/?/init.lua"

	l := lua.NewState(lua.Options{FS: &fs.FS{EmbedFS: goks.LuaTree}})
	l.PreloadModule("go.json", json.Loader)
	l.PreloadModule("go.rand", rand.Loader)
	l.PreloadModule("go.uuid", uuid.Loader)
	l.PreloadModule("go.ipmatcher", ipmatcher.Loader)
	l.PreloadModule("go.re2", gluare.Loader)
	ngx.LoadNgx(l)

	if err := setup(l); err != nil {
		return nil, err
	}
	return &VM{l: l}, nil
}

func (v *VM) Close() {
	v.l.Close()
}

func (v *VM) Execute(file io.Reader, filename string) error {
	l := v.l
	fn, err := l.Load(file, filename)
	if err != nil {
		return err
	}
	l.Push(fn)
	err = l.PCall(0, lua.MultRet, nil)
	return err
}

func (v *VM) CallByParams(name string, args ...string) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	luaArgs := make([]lua.LValue, 0, len(args))
	for _, arg := range args {
		luaArgs = append(luaArgs, lua.LString(arg))
	}
	argCount := 2
	l := v.l
	err := l.CallByParam(lua.P{
		Fn:      l.GetGlobal(name),
		NRet:    argCount,
		Protect: true,
		Handler: nil,
	}, luaArgs...)
	if err != nil {
		return "", err
	}
	var (
		resS string
		resE error
	)
	if ret1 := l.Get(-2); ret1 != nil {
		resS = lua.LVAsString(ret1)
	}
	if ret2 := l.Get(-1); ret2 != nil {
		errString := lua.LVAsString(ret2)
		if errString != "" {
			resE = fmt.Errorf(errString)
		}
	}
	l.Pop(argCount)
	return resS, resE
}
