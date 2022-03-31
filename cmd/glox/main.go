package main

import (
	"os"

	"github.com/iamsayantan/glox"
)

func main() {
	args := os.Args[1:]

	runtime := glox.NewRuntime()
	runtime.Run(args)
}
