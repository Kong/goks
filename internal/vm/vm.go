package vm

import (
	"io"

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
	l *lua.LState
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

func (v *VM) Execute(file io.Reader, filename string) (string, error) {
	l := v.l
	fn, err := l.Load(file, filename)
	if err != nil {
		return "", err
	}
	l.Push(fn)
	err = l.PCall(0, lua.MultRet, nil)
	if err != nil {
		return "", err
	}
	return "", nil
}
