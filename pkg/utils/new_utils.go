package utils

import (
	"fmt"
)

func NewFunc() error {
	for i := 0; i < 10; i++ {
		fmt.Printf("Dummy function %d", i)
	}
	//// this would cause lint issues
	//switch _, e := strconv.ParseBool("true"); e.(type) {
	//case errorutils.CodedError:
	//	return nil
	//}
	return nil
}
