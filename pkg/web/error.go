package web

import (
	"fmt"
	"reflect"
)

const (
	templateInvalidMvcHandlerFunc = "invalid MVC handler function signature: %v, but got <%v>"
)

type errorInvalidMvcHandlerFunc struct {
	reason error
	target *reflect.Value
}

func (e *errorInvalidMvcHandlerFunc) Error() string {
	return fmt.Sprintf(templateInvalidMvcHandlerFunc, e.reason.Error(), e.target.Type())
}
