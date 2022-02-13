package log

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func Test_GlobalLog(t *testing.T) {
	buffer := &bytes.Buffer{}
	SetLogger(NewStdLogger(buffer))

	testCases := []struct{
		level   Level
		content []interface{}
	}{
		{
			level: LevelError,
			content: []interface{}{"test error"},
		},
	}

	expected := make([]string, 0)
	for _, tc := range testCases {
		msg := fmt.Sprintf(tc.content[0].(string), tc.content[1:]...)
		switch tc.level {
		case LevelError:
			Errorf(tc.content[0].(string), tc.content[1:]...)
			expected = append(expected, fmt.Sprintf("%v msg=%s", "ERROR", msg))
		}
	}
	expected = append(expected, "")
	t.Logf("Content: %v", buffer.String())
	if buffer.String() != strings.Join(expected, "\n") {
		t.Errorf("Expected: %v, got: %v", strings.Join(expected, "\n"), buffer.String())
	}
}