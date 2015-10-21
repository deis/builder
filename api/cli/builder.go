package main

import (
	"os"
	"runtime"

	"github.com/deis/deis/buildpack/builder"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	os.Exit(builder.Run("boot"))
}
