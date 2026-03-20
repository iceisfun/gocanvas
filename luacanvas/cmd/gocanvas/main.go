package main

import (
	"fmt"
	"log"
	"os"

	"github.com/iceisfun/gocanvas/luacanvas"
	"github.com/iceisfun/golua/compiler"
	"github.com/iceisfun/golua/parser"
	"github.com/iceisfun/golua/stdlib"
	"github.com/iceisfun/golua/vm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <script.lua>\n", os.Args[0])
		os.Exit(1)
	}

	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	block, err := parser.Parse(os.Args[1], string(source))
	if err != nil {
		log.Fatal(err)
	}

	proto, err := compiler.Compile(os.Args[1], block)
	if err != nil {
		log.Fatal(err)
	}

	v := vm.New()
	stdlib.Open(v)

	b := luacanvas.New()
	b.Register(v)

	if _, err := v.Run(proto); err != nil {
		log.Fatal(err)
	}
}
