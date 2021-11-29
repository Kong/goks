package main

import (
	"flag"
	"fmt"
	"os"

	json "github.com/layeh/gopher-json"
	lua "github.com/yuin/gopher-lua"
)

func main() {
	os.Exit(mainAux())
}

func mainAux() int {
	flag.Parse()

	L := lua.NewState(lua.Options{})
	defer L.Close()
	L.SetMx(1)
	L.PreloadModule("json", json.Loader)

	if nargs := flag.NArg(); nargs > 0 {
		script := flag.Arg(0)
		if err := execScript(L, script); err != nil {
			fmt.Println(err.Error())
			return 1
		}
	}
	return 0
}

func execScript(L *lua.LState, script string) error {
	file, err := os.Open(script)
	if err != nil {
		return err
	}
	defer file.Close()
	fn, err := L.Load(file, script)
	if err != nil {
		return err
	}
	L.Push(fn)
	err = L.PCall(0, lua.MultRet, nil)
	if err != nil {
		return err
	}
	ret := L.Get(1)
	fmt.Println(ret)
	L.Pop(1)
	return nil
}
