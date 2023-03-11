package log

import (
	"fmt"
	"testing"
)

func TestError(t *testing.T) {
	err := ErrLockedUser(nil, "thanh", M{"number2": "thanh"})

	fmt.Println("EEEEEE", err.Error())

	err = ErrNotFound(nil, "thanh", "user", M{"number2": 40})

	fmt.Println("EEEEEE", err.Error())

	err = ErrServer(nil, err, M{"number2": 40})

	fmt.Println("EEEEEE", err.Error())
}
