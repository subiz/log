package log

import (
	"encoding/json"
	"os"
	"testing"
)

func TestFormat(t *testing.T) {
	var cases []struct {
		Name     string
		Template string
		Data     map[string]any
		Expected string
	}
	data, err := os.ReadFile("testdata/format.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			if got := formatString(tc.Template, tc.Data); got != tc.Expected {
				t.Errorf("format(%q) = %q, want %q", tc.Template, got, tc.Expected)
			}
		})
	}
}
