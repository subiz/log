package log_test

import (
	"fmt"
	"testing"
	log "github.com/subiz/log"
)

func A() error {
	err :=  B()
	return err
}










func B() error {
	err :=  CCCCCC()
	return err
}


func CCCCCC() error {
	err :=  DDDDDD()
	return err
}


func DDDDDD() error {
	err := EEEE()
	return err
}






func EEEE() error {
	err :=  log.EData(nil, nil, log.M{"number2": "thanh"})
	return err
}

func TestError(t *testing.T) {
	err := A()
	fmt.Println("EEEEEE", err.Error())
}
