package svn

//#include <svn_types.h>
//#include <svn_pools.h>
//#include <svn_fs.h>
//#include <svn_delta.h>
//#include <svn_repos.h>
//extern void init_fs_config(apr_hash_t *fsConfig);
import "C"

import (
	"unsafe"
)

// Create a repository
// If we will copy from the template, set new uuid: err := C.svn_fs_set_uuid(r.fs, nil, r.pool)
func Create(path string) error {
	var (
		err   *C.svn_error_t
		repos *C.svn_repos_t
	)

	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	pool := initSubPool(globalPool)
	defer C.svn_pool_destroy(pool)

	var fsConfig *C.apr_hash_t = C.apr_hash_make(pool)
	C.init_fs_config(fsConfig)
	err = C.svn_repos_create(&repos, cstr, nil, nil, nil, fsConfig, pool)

	if err != nil {
		return makeError(err)
	}

	return nil
}
