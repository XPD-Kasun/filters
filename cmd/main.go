package main

import (
	"fmt"
	"os"
)

func main() {

	args := os.Args[1:]
	fmt.Printf("%+v\n", args)
	gofile := os.Getenv("GOFILE")
	goline := os.Getenv("GOLINE")
	gopkg := os.Getenv("GOPACKAGE")
	fmt.Println(gofile, goline, gopkg)

	ctx := &GenCtx{
		line: goline,
		path: gofile,
		pkg:  gopkg,
	}

	targetFileName := args[0]
	if targetFileName == "" {
		targetFileName = gofile + "_query.go"
	}
	gen := NewGenerator(ctx)

	gen.Generate()
	fmt.Println(gen.GetCode())
	gen.Sync(targetFileName)

	// fs := token.NewFileSet()
	// parser.Mode

}
