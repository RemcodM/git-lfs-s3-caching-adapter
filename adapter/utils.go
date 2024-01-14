package adapter

import (
	"fmt"
	"os"
)

// standaloneFailure reports a fatal error.
func standaloneFailure(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err)
	os.Exit(2)
}
