package testutils

import (
	"os"
	"strings"
)

type testEnvironment struct {	
	envMap map[string]string
}

func NewTestEnvironment() *testEnvironment {
	te := &testEnvironment{
		envMap: make(map[string]string),
	}

	for _,pair := range os.Environ() {
		parts := strings.SplitN(pair, "=", 2)
		te.envMap[parts[0]] = parts[1]
	}

	return te
}

func (te *testEnvironment) Getenv(key string) string {
	return te.envMap[key]
}

func (te *testEnvironment) Setenv(key, val string) (err error) {
	if strings.Contains(key, "=") {
		err = os.ErrInvalid
		return
	}
	te.envMap[key] = val
	return
}

func (te *testEnvironment) Unsetenv(key string) (err error) {
	if strings.Contains(key, "=") {
		err = os.ErrInvalid
		return
	}
	delete(te.envMap, key)
	return
}

func (te *testEnvironment) Environ() (list []string) {
	list = make([]string, 0, len(te.envMap))
	for k,v := range te.envMap {
		list = append(list, k + "=" + v)
	}
	return
}

func (te *testEnvironment) Clearenv() {
	te.envMap = make(map[string]string)
}

func (te *testEnvironment) LookupEnv(key string) (string, bool) {
	val, exists := te.envMap[key]
	return val, exists
}

func (te *testEnvironment) ExpandEnv(input string) string {
	return os.Expand(input, te.Getenv)
}
