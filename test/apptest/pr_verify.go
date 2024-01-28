package apptest

import "fmt"

func NewFunc() error {
	for i := 0; i < 10; i++ {
		fmt.Printf("Dummy function %d", i)
	}
	return nil
}