package main

import "fmt"

const usage = `
All services should provide a top-level Go program named "boot.go", compiled
to "boot". This should go in rootfs/bin/boot, and should be the entry point
for all Deis components.

An exception to this is components that may be started without a boot, or which
may be started with a simple (<20 line) shell script.
`

func main() {
	fmt.Println(usage)
}
