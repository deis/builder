package main

import (
	"os"
	"runtime"

	"github.com/deis/builder/pkg"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	os.Exit(pkg.Run("boot"))
}
