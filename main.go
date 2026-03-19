package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gox <compile|check> <file.gox>")
		os.Exit(1)
	}
	fmt.Println("gox: not yet implemented")
}
