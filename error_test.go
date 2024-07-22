package log_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	log "github.com/subiz/log"
)

func A() error {
	err := B()
	return err
}

func B() error {
	err := CCCCCC()
	return err
}

func CCCCCC() error {
	err := DDDDDD()
	return err
}

func DDDDDD() error {
	err := E()
	return err
}

func E() error {
	err := log.EServer(nil, log.M{"number3": "thanh"})
	return err
}

func TestError(t *testing.T) {
	// err := log.EInvalidZaloToken("thanh", "3290323", "Dayladau") // A()
	err := A()
	fmt.Println("EEEEEE", err.Error())
}

func TestLogErr(t *testing.T) {
	err := log.EInvalidGoogleToken("thanh", "3290323", "Dayladau") // A()
	log.Err("subiz", err, "param")
}

func TestWrap(t *testing.T) {
	var err error = log.EAccountLocked("thanh") // A()
	log.WrapStack(err, 0)
	time.Sleep(20 * time.Second)
}

func TestToJson(t *testing.T) {
	err := log.EAccountLocked("thanh") // A()
	str := err.ToJSON()
	fmt.Println("JSON:", str)
}

func TestMarshalJson(t *testing.T) {
	err := log.EAccountLocked("thanh") // A()
	str := log.NewError(err, nil).ToJSON()
	fmt.Println("JSON--------:", str)
}

func TestDontModify(t *testing.T) {
	err := log.EAccountLocked("thanh") // A()

	log.EServer(err, log.M{"secret": "2340asdffEIFJ42"})
	str := log.NewError(err, nil).ToJSON()
	if strings.Contains(str, "2340asdffEIFJ42") {
		t.Error("should not contain")
	}
}
