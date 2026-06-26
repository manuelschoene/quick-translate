package main

import (
	"fmt"
	"os"
)

func main() {
	cp, err := ClipboardInit()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	out, err := cp.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(out)
}