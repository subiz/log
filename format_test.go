package log

import (
	"testing"
)

func TestFormat(t *testing.T) {
	flagtests := []struct {
		template string
		data     map[string]interface{}
		out      string
	}{
		{"hi {name}", map[string]interface{}{
			"name": "thanh",
		}, "hi thanh"},
		{"hi {name}", map[string]interface{}{
			"name": "{name}thanh",
		}, "hi {name}thanh"},
		{"hi {{name{name}}", map[string]interface{}{
			"name": "thanh",
		}, "hi {namethanh}"},
		{"conversation_started.account.{account_id}.user.{user_id}", map[string]interface{}{
			"account_id": "1",
			"user_id":    "2",
			"client_id":  "3",
		}, "conversation_started.account.1.user.2"},
		{"hi {name2}", nil, "hi {name2}"},
		{"hi {{name2} alo", nil, "hi {name2} alo"},
		{"hi {{name2", nil, "hi {name2"},
		{"hi {name2 {name}", map[string]interface{}{
			"name": "thanh",
		}, "hi {name2 thanh"},
		{"hi {name2 {name}$s", map[string]interface{}{
			"name": "thanh",
		}, "hi {name2 thanh$s"},
		{"hi {name2#{name$s", map[string]interface{}{
			"name": "thanh",
		}, "hi {name2#{name$s"},
		{"hi {name{name}$v", map[string]interface{}{
			"name": "thanh",
		}, "hi {namethanh$v"},
	}

	for _, tt := range flagtests {
		s := formatString(tt.template, tt.data)
		if s != tt.out {
			t.Errorf("should equal, input %s expect %s, got %s", tt.template, tt.data, s)
		}
	}
}
