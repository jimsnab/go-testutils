package testutils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func ExpectFileExists(io FileIo, t *testing.T, filePath string) {
	t.Helper()

	fi, err := io.Stat(filePath)
	ExpectError(t, nil, err)
	ExpectBool(t, false, fi.IsDir())
}

func ExpectFileJson(io FileIo, t *testing.T, filePath string, data map[string]interface{}) {
	t.Helper()

	fi, err := io.Stat(filePath)
	ExpectError(t, nil, err)
	ExpectBool(t, false, fi.IsDir())

	content, err := io.ReadFile(filePath)
	ExpectError(t, nil, err)

	var contentJson map[string]interface{}
	err = json.Unmarshal(content, &contentJson)
	ExpectError(t, nil, err)

	dbg, _ := json.MarshalIndent(contentJson, "", "  ")
	fmt.Printf("%s\n", string(dbg))

	ExpectBool(t, true, reflect.DeepEqual(data, contentJson))
}

func isDataStored(context string, allData interface{}, part interface{}) bool {
	if reflect.TypeOf(allData) != reflect.TypeOf(part) {
		fmt.Printf("%s has different types\n", context)
		return false
	}

	switch val := part.(type) {
	case []interface{}:
		allArray := allData.([]interface{})
		if len(allArray) != len(val) {
			fmt.Printf("%s array lengths are different\n", context)
			return false
		}
		for pos, item := range val {
			if !isDataStored(fmt.Sprintf("%s[%d]", context, pos), allArray[pos], item) {
				return false
			}
		}
		return true

	case map[string]interface{}:
		allMap := allData.(map[string]interface{})
		for k, v := range val {
			allV, exist := allMap[k]
			if !exist || !isDataStored(fmt.Sprintf("%s.%s", context, k), allV, v) {
				return false
			}
		}
		return true

	default:
		if !reflect.DeepEqual(allData, part) {
			fmt.Printf("%s has different values \"%v\" vs \"%v\"\n", context, allData, part)
			return false
		}
		return true
	}
}

func ExpectFileJsonPart(io FileIo, t *testing.T, filePath string, subsetData map[string]interface{}) {
	t.Helper()

	fi, err := io.Stat(filePath)
	ExpectError(t, nil, err)
	ExpectBool(t, false, fi.IsDir())

	content, err := io.ReadFile(filePath)
	ExpectError(t, nil, err)

	var fullJson map[string]interface{}
	err = json.Unmarshal(content, &fullJson)
	ExpectError(t, nil, err)

	dbg, _ := json.MarshalIndent(fullJson, "", "  ")
	fmt.Printf("%s\n", string(dbg))

	ExpectBool(t, true, isDataStored("root", fullJson, subsetData))
}
