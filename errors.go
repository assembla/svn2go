package svn

// #include <svn_types.h>
import "C"

import (
	"fmt"
)

// Error is used for wrapping SVN errors
type Error struct {
	err *C.svn_error_t
}

func makeError(err *C.svn_error_t) error {
	defer C.svn_error_clear(err)
	return fmt.Errorf("%s", C.GoString(err.message))
}
