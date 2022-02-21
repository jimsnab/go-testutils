package testutils

import (
	"encoding/json"
	"fmt"
	"strings"
)

func NiceJs(testJson string) string {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(testJson), &data)
	if err != nil {
		panic(fmt.Errorf("invalid test json"))
	}
	result, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(fmt.Errorf("unexpected json marshal error"))
	}
	return string(result)
}

func NiceYaml(testYaml string) string {
	const tabStop = 2

	var sb strings.Builder
	col := 0
	for _,ch := range testYaml {
		if ch == '\n' {
			col = 0
		} else if ch == '\t' {
			sb.WriteRune(' ')
			col++
			for col % tabStop != 0 {
				sb.WriteRune(' ')
				col++
			}
			continue
		} else if (ch < 32) {
			continue
		} else {
			col++
		}
		sb.WriteRune(ch)
	}

	testYaml = sb.String()
	maxSpaces := len(testYaml)

	lines := strings.Split(testYaml, "\n")
	for _,line := range lines {
		if len(line) > 0 {
			spaces := 0
			for spaces < len(line) && line[spaces] == ' ' {
				spaces++
			}

			if spaces < maxSpaces && spaces < len(line) {
				maxSpaces = spaces
			}
		}
	}

	sb.Reset()
	for _, line := range lines {
		if len(line) > maxSpaces {
			text := line[maxSpaces:]
			if len(text) > 0 {
				sb.WriteString(text + "\n")
			}
		}
	}

	return sb.String()
}
