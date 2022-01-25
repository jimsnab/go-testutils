package testutils

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/jimsnab/go-simpleutils"
)

var escape = simpleutils.Escape

func ExpectPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		t.Helper()
		r := recover()
		if r == nil {
			t.Errorf("Did not get expected panic")
		}
	}()

	fn()
}

func ExpectPanicError(t *testing.T, expectedError error, fn func()) {
	t.Helper()
	defer func() {
		t.Helper()
		r := recover()
		if r == nil {
			if expectedError != nil {
				t.Errorf("Did not get expected panic. Type expected: %T", expectedError)
			}
		} else {
			if reflect.TypeOf(r) != reflect.TypeOf(expectedError) {
				t.Errorf("Got error type \"%s\" but expected \"%s\"", reflect.TypeOf(r), reflect.TypeOf(expectedError))
			} else {
				testError := r.(error)
				if expectedError.Error() != testError.Error() {
					t.Errorf("Got %T \"%v\" but expected \"%v\"", testError, testError, expectedError)
				}
			}
		}
	}()

	fn()
}

func DoMapsMatch(t *testing.T, expectedMap map[string]interface{}, testMap map[string]interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expectedMap, testMap) {
		t.Errorf("Did not get expected %v, got %v", expectedMap, testMap)
	}
}

func ExpectError(t *testing.T, expectedError error, testError error) {
	t.Helper()
	if expectedError == nil {
		if testError != nil {
			t.Errorf("Got unexpected %T: \"%v\"", testError, testError)
		}
	} else {
		if testError == nil {
			t.Errorf("Did not get expected %T: \"%v\"", expectedError, expectedError)
		} else if reflect.TypeOf(expectedError) != reflect.TypeOf(testError) {
			t.Errorf("Got error type \"%s\" but \"%s\" was expected", reflect.TypeOf(testError), reflect.TypeOf(expectedError))
		} else if expectedError.Error() != testError.Error() {
			t.Errorf("Got %T\n \"%v\"\nbut\n \"%v\"\nwas expected", testError, testError, expectedError)
		}
	}
}

func ExpectErrorContainingText(t *testing.T, expectedSubtext string, testError error) {
	t.Helper()
	if testError == nil {
		t.Errorf("Did not get expected error")
	} else if !strings.Contains(testError.Error(), expectedSubtext) {
		t.Errorf("Got %T\n \"%v\"\nbut\n \"%s\"\nwas expected", testError, testError, expectedSubtext)
	}
}

func ExpectBool(t *testing.T, expectedBool bool, testBool bool) {
	t.Helper()
	if expectedBool != testBool {
		t.Errorf("Got %v but %v expected", testBool, expectedBool)
	}
}

func ExpectString(t *testing.T, expectedStr string, testStr string) {
	t.Helper()
	if expectedStr != testStr {
		t.Errorf("Got\n \"%s\"\nbut\n \"%s\"\nwas expected", escape(testStr), escape(expectedStr))
	}
}

func ExpectValue(t *testing.T, expectedVal interface{}, testVal interface{}) {
	t.Helper()
	if expectedVal != testVal {
		t.Errorf("Got %v but %v expected", testVal, expectedVal)
	}
}

func CaptureStdout(t *testing.T, fn func()) string {
	orgStdout := os.Stdout
	defer func() { os.Stdout = orgStdout }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
		return ""
	}

	os.Stdout = w

	output := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil {
			t.Error(err)
			output <- ""
		} else {
			output <- buf.String()
		}
	}()

	fn()
	w.Close()

	result := <-output
	return result
}
