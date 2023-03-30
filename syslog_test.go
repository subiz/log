package log

import (
	"errors"
	"testing"
)

func TestSyslog(t *testing.T) {
	if err := Info("mothaiba", "thanh", []byte("dao")); err != nil {
		println(err)
		t.Fatal(err)
	}
}

func a() {
	b()
}

func b() {
	c()
}

func c() {
	var a = 10
	a++
	Err("mothaiba", errors.New("database_error"), "thanh")
}
func TestSyslogError(t *testing.T) {
	a()
}
