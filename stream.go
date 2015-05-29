package svn

/*
#cgo pkg-config: apr-1
#include <svn_io.h>
#include <svn_pools.h>
*/
import "C"

import (
	"io"
	"runtime"
	"unsafe"
)

// TODO svn stream supports seek, maybe we can use it
type SvnStream struct {
	io   *C.svn_stream_t
	pool *C.apr_pool_t
}

// Read bytes from svn stream
func (s *SvnStream) Read(dest []byte) (n int, err error) {
	c := C.apr_size_t(len(dest))

	if err := C.svn_stream_read(s.io, (*C.char)(unsafe.Pointer(&dest[0])), &c); err != nil {
		return int(c), makeError(err)
	}

	if c == 0 {
		return 0, io.EOF
	} else {
		return int(c), nil
	}
}

// Closes svn stream
func (s *SvnStream) Close() error {
	runtime.SetFinalizer(s, nil)

	err := C.svn_stream_close(s.io)
	C.svn_pool_destroy(s.pool)

	if err != nil {
		return makeError(err)
	}

	return nil
}
