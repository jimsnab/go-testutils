package testutils

import "os"

type processEnvironment interface {
	Getenv(key string) string
	Setenv(key, val string) (err error)
	Unsetenv(key string) (err error)
	Environ() []string
	Clearenv()
	LookupEnv(key string) (string, bool)
	ExpandEnv(input string) string
}

type realEnvironment struct {	
}

var Environ = processEnvironment(&realEnvironment{})


func (re *realEnvironment) Getenv(key string) string {
	return os.Getenv(key)
}

func (re *realEnvironment) Setenv(key, val string) (err error) {
	err = os.Setenv(key, val)
	return
}

func (re *realEnvironment) Unsetenv(key string) (err error) {
	err = os.Unsetenv(key)
	return
}

func (re *realEnvironment) Environ() []string {
	return os.Environ()
}

func (re *realEnvironment) Clearenv() {
	os.Clearenv()
}

func (re *realEnvironment) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func (re *realEnvironment) ExpandEnv(input string) string {
	return os.ExpandEnv(input)
}