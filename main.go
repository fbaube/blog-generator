package main

import (
	// "fmt"
	"github.com/fbaube/bloggenator/cli"
	// "github.com/morningconsult/serrors"
)

// main is a dead simple one-liner eh ?
func main() {
	/*
	s := serrors.New("This is a stack trace error (only \"+v\" works)")
	fmt.Printf("====\n%+v\n====\n",s)
	*/
	cli.Run()
}
