package main

import (
	"fmt"
	"os"

	"gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/adapter"
)

func main() {
	err := adapter.ProcessData(os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
}
