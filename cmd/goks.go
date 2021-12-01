package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kong/goks/internal/vm"
)

func main() {
	os.Exit(mainAux())
}

func mainAux() int {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("please provide a script to run")
		return 1
	}
	script := flag.Arg(0)

	luaVM, err := vm.New()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	file, err := os.Open(script)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	defer file.Close()

	ret, err := luaVM.Execute(file, script)
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}
	fmt.Println(ret)
	return 0
}
