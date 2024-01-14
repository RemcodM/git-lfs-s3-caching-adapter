package lfs

import (
	"fmt"
	"os"
	"strings"

	"github.com/git-lfs/git-lfs/v3/config"
)

type passthroughEnvironment struct {
	environment config.Environment
}

func newPassthroughEnvironment(environment config.Environment) config.Environment {
	return &passthroughEnvironment{environment}
}

func (e *passthroughEnvironment) Get(key string) (string, bool) {
	value, ok := e.environment.Get(key)
	if key == "lfs.standalonetransferagent" && ok && value == "caching" {
		fmt.Fprintf(os.Stderr, "Call to read %s returned %s, intercepting and returning empty value\n", key, value)
		return "", false
	}
	if key == "lfs.url" && ok && strings.HasPrefix(value, "caching::") {
		url := strings.TrimPrefix(value, "caching::")
		if url == "" {
			fmt.Fprintf(os.Stderr, "Call to read %s returned %s, intercepting and returning empty value\n", key, value)
			return "", false
		}
		fmt.Fprintf(os.Stderr, "Call to read %s returned %s, intercepting and returning %s\n", key, value, url)
		return url, true
	}
	return value, ok
}

func (e *passthroughEnvironment) GetAll(key string) []string {
	return e.environment.GetAll(key)
}

func (e *passthroughEnvironment) Bool(key string, def bool) bool {
	return e.environment.Bool(key, def)
}

func (e *passthroughEnvironment) Int(key string, def int) int {
	return e.environment.Int(key, def)
}

func (e *passthroughEnvironment) All() map[string][]string {
	return e.environment.All()
}
