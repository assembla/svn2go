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

// Stream is used for wrapping SVN stream
// TODO svn stream supports seek, maybe we can use it
type Stream struct {
	io   *C.svn_stream_t
	pool *C.apr_pool_t
}

// Read bytes from svn stream
func (s *Stream) Read(dest []byte) (n int, err error) {
	c := C.apr_size_t(len(dest))

	if err := C.svn_stream_read_full(s.io, (*C.char)(unsafe.Pointer(&dest[0])), &c); err != nil {
		return int(c), makeError(err)
	}

	if c == 0 {
		return 0, io.EOF
	}
	return int(c), nil
}

// Close closes stream
func (s *Stream) Close() error {
	runtime.SetFinalizer(s, nil)
	C.svn_pool_destroy(s.pool)
	return nil
}
