package svn

/*
#cgo LDFLAGS: -lsvn_delta-1 -lsvn_repos-1 -lsvn_subr-1 -lsvn_fs-1
#cgo LDFLAGS: -lsvn_fs_util-1 -lsvn_fs_fs-1 -lsvn_diff-1

#cgo LDFLAGS: -L/usr/local/lib -L/usr/local/opt/subversion/lib -L/usr/lib
#cgo CFLAGS: -I/usr/local/include/subversion-1 -I/usr/local/opt/subversion/include/subversion-1 -I/usr/include/subversion-1

#cgo darwin LDFLAGS: -framework CoreFoundation -framework Security -framework CoreServices -g

#cgo pkg-config: apr-1 apr-util-1

#include <svn_hash.h>
#include <svn_pools.h>
#include <svn_path.h>
#include <svn_repos.h>
#include <svn_error.h>
#include <svn_dso.h>
#include <svn_utf.h>
#include <svn_props.h>

extern svn_error_t * FileMimeType(svn_string_t **mimetype, svn_fs_root_t *root, const char *path, apr_pool_t *pool);
extern apr_array_header_t * GoCreateAprArrayForPath(const char *path, apr_pool_t *pool);
extern apr_array_header_t * GoDefaultLogProps(apr_pool_t *pool);
extern svn_error_t * Go_svn_repos_get_logs4(svn_repos_t *repos,
                    const apr_array_header_t *paths,
                    svn_revnum_t start,
                    svn_revnum_t end,
                    int limit,
                    svn_boolean_t discover_changed_paths,
                    svn_boolean_t strict_node_history,
                    svn_boolean_t include_merged_revisions,
                    const apr_array_header_t *revprops,
                    void *receiver_baton,
                    apr_pool_t *pool);
extern svn_error_t * Go_repos_history(svn_fs_t *fs,
                   const char *path,
                   void *history_baton,
                   svn_revnum_t start,
                   svn_revnum_t end,
                   svn_boolean_t cross_copies,
                   apr_pool_t *pool);
*/
import "C"

import (
	"io"
	"log"
	"runtime"
	"unsafe"
)

// global pool for svn errors and other libraries data
var globalPool *C.apr_pool_t

func init() {
	C.apr_initialize()
	//C.atexit(C.apr_terminate) TODO FIXME, maybe we can set finalizer to globalPool?
	err := C.svn_dso_initialize2()

	if err != nil {
		log.Fatal(makeError(err))
	}

	C.svn_utf_initialize2(C.FALSE, C.svn_pool_create_ex(nil, nil))
	globalPool = C.apr_allocator_owner_get(C.svn_pool_create_allocator(C.TRUE))

	// from svnlook command
	/* Initialize the FS library. */
	err = C.svn_fs_initialize(globalPool)

	if err != nil {
		log.Fatal(makeError(err))
	}
}

// Opens svn repository
func Open(path string) (*Repo, error) {
	r := &Repo{Path: path}
	r.pool = C.svn_pool_create_ex(globalPool, nil)
	runtime.SetFinalizer(r, (*Repo).Close)
	cs := C.CString(path) // convert to C string
	defer C.free(unsafe.Pointer(cs))

	p := C.svn_dirent_internal_style(cs, r.pool)

	err := C.svn_repos_open2(&r.repos, p, nil, r.pool)

	if err != nil {
		r.Close() // free pool
		return nil, makeError(err)
	}

	r.fs = C.svn_repos_fs(r.repos)

	return r, nil
}

func (r *Repo) LatestRevision() (int64, error) {
	var (
		rev C.svn_revnum_t
	)

	subPool := initSubPool(r.pool)
	defer C.svn_pool_destroy(subPool)

	if err := C.svn_fs_youngest_rev(&rev, r.fs, subPool); err != nil {
		return 0, makeError(err)
	}

	revC := int64(rev)
	// copy data
	rev64 := revC
	return rev64, nil
}

func (r *Repo) CommitInfo(rev int64) (*Commit, error) {
	r.rev = rev
	c := &Commit{Rev: rev}
	var err error
	c.Author, err = r.GetProperty(PropRevAuthor)

	if err != nil {
		return nil, err
	}

	c.Date, err = r.GetProperty(PropRevDate)

	if err != nil {
		return nil, err
	}

	c.Log, err = r.GetProperty(PropRevLog)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (r *Repo) GetProperty(name string) (string, error) {
	var (
		rawValue *C.svn_string_t
	)

	propName := C.CString(name)
	defer C.free(unsafe.Pointer(propName))

	subPool := initSubPool(r.pool)
	defer C.svn_pool_destroy(subPool)

	err := C.svn_fs_revision_prop(&rawValue, r.fs, C.svn_revnum_t(r.rev), propName, subPool)

	if err != nil {
		return "", makeError(err)
	}

	prop := C.GoString(rawValue.data)
	return prop, nil
}

func (r *Repo) PropGet(path string, rev int64, propName string) (string, error) {
	var (
		value *C.svn_string_t
	)

	result := ""
	propname := C.CString(propName)
	defer C.free(unsafe.Pointer(propname))

	target := C.CString(path)
	defer C.free(unsafe.Pointer(target))

	var revisionRoot *C.svn_fs_root_t

	subPool := initSubPool(r.pool)
	defer C.svn_pool_destroy(subPool)

	if err := C.svn_fs_revision_root(
		&revisionRoot,
		r.fs,
		C.svn_revnum_t(rev),
		subPool,
	); err != nil {
		return "", makeError(err)
	}

	// svn_error_t *
	// svn_fs_node_prop(svn_string_t **value_p,
	//                  svn_fs_root_t *root,
	//                  const char *path,
	//                  const char *propname,
	//                  apr_pool_t *pool);
	if err := C.svn_fs_node_prop(
		&value,
		revisionRoot,
		target,
		propname,
		subPool,
	); err != nil {
		return "", makeError(err)
	}

	if value != nil {
		result = C.GoString(value.data)
	}

	return result, nil
}

func (r *Repo) PropList(path string, rev int64) (map[string]string, error) {
	var (
		revisionRoot *C.svn_fs_root_t
		props        *C.apr_hash_t
	)

	target := C.CString(path)
	defer C.free(unsafe.Pointer(target))

	subPool := initSubPool(r.pool)
	defer C.svn_pool_destroy(subPool)

	if err := C.svn_fs_revision_root(
		&revisionRoot,
		r.fs,
		C.svn_revnum_t(rev),
		subPool,
	); err != nil {
		return nil, makeError(err)
	}

	// svn_error_t*
	// svn_fs_node_proplist(apr_hash_t **table_p,
	//                      svn_fs_root_t *root,
	//                      const char *path,
	//                      apr_pool_t *pool);

	if err := C.svn_fs_node_proplist(&props,
		revisionRoot,
		target,
		subPool,
	); err != nil {
		return nil, makeError(err)
	}

	rez := make(map[string]string)

	for hi := C.apr_hash_first(subPool, props); hi != nil; hi = C.apr_hash_next(hi) {
		var (
			key, val unsafe.Pointer
		)

		C.apr_hash_this(hi, &key, nil, &val)
		propKey := C.GoString((*C.char)(key))
		propVal := (*C.svn_string_t)(val)

		rez[propKey] = C.GoString(propVal.data)
	}

	return rez, nil
}

func (r *Repo) Close() error {
	C.svn_pool_destroy(r.pool)
	runtime.SetFinalizer(r, nil)
	return nil
}

// List repository revisions
func (r *Repo) Commits(start, end int64) ([]Commit, error) {
	subPool := initSubPool(r.pool)
	baton := NewCommitCollector(subPool)

	err := C.Go_svn_repos_get_logs4(r.repos,
		nil,
		C.svn_revnum_t(start),
		C.svn_revnum_t(end),
		20,      // limit
		C.FALSE, //discover_changed_paths
		C.FALSE, // strict_node_history
		C.FALSE, // include_merged_revisions
		C.GoDefaultLogProps(baton.pool),
		unsafe.Pointer(baton),
		baton.pool)

	if err != nil {
		return nil, makeError(err)
	}

	return baton.commits, nil
}

// Returns commits that changed @path
func (r *Repo) History(path string, start, end int64, limit int) ([]Commit, error) {
	cpath := C.CString(path) // convert to C string
	defer C.free(unsafe.Pointer(cpath))

	subPool := initSubPool(r.pool)
	baton := NewCommitCollector(subPool)

	baton.limit = limit
	baton.r = r

	err := C.Go_repos_history(r.fs, cpath, unsafe.Pointer(baton),
		C.svn_revnum_t(start),
		C.svn_revnum_t(end),
		C.TRUE, // cross copies?
		subPool)
	if err != nil {
		return nil, makeError(err)
	}

	return baton.commits, nil
}

// Return an array with directory entries
func (r *Repo) Tree(path string, rev int64) ([]DirEntry, error) {
	var revisionRoot *C.svn_fs_root_t

	subPool := initSubPool(r.pool)
	defer C.svn_pool_destroy(subPool)

	if err := C.svn_fs_revision_root(&revisionRoot, r.fs, C.svn_revnum_t(rev), subPool); err != nil {
		return nil, makeError(err)
	}

	var entries *C.apr_hash_t
	cpath := C.CString(path) // convert to C string
	defer C.free(unsafe.Pointer(cpath))

	if err := C.svn_fs_dir_entries(&entries, revisionRoot, cpath, subPool); err != nil {
		return nil, makeError(err)
	}

	rez := make([]DirEntry, C.apr_hash_count(entries))
	i := 0

	for hi := C.apr_hash_first(subPool, entries); hi != nil; hi = C.apr_hash_next(hi) {
		var (
			entry *C.svn_fs_dirent_t
			val   unsafe.Pointer
			kind  int
		)

		C.apr_hash_this(hi, nil, nil, &val)
		entry = (*C.svn_fs_dirent_t)(val)
		k := int(entry.kind)
		// copy data
		kind = k
		rez[i] = DirEntry{C.GoString(entry.name), kind}
		i++
	}

	return rez, nil
}

// Returns file as ReaderCloser stream
func (r *Repo) FileContent(path string, rev int64) (io.ReadCloser, error) {
	var (
		revisionRoot *C.svn_fs_root_t
	)

	subPool := initSubPool(r.pool)

	if err := C.svn_fs_revision_root(&revisionRoot, r.fs, C.svn_revnum_t(rev), subPool); err != nil {
		return nil, makeError(err)
	}

	return r.initSvnStream(revisionRoot, subPool, path)
}

// Returns latest revision for the path
func (r *Repo) LastPathRev(path string, baseRev int64) (int64, error) {
	var (
		revisionRoot *C.svn_fs_root_t
		rev          C.svn_revnum_t
	)
	subPool := initSubPool(r.pool)
	defer C.svn_pool_destroy(subPool)

	if e := C.svn_fs_revision_root(&revisionRoot, r.fs, C.svn_revnum_t(baseRev), subPool); e != nil {
		return -1, makeError(e)
	}

	cpath := C.CString(path) // convert to C string
	defer C.free(unsafe.Pointer(cpath))

	if err := C.svn_fs_node_created_rev(&rev, revisionRoot, cpath, subPool); err != nil {
		return -1, makeError(err)
	}

	revC := int64(rev)
	// copy data
	rev64 := revC
	return rev64, nil
}

// Returns file size
func (r *Repo) FileSize(path string, rev int64) (int64, error) {
	var (
		revisionRoot *C.svn_fs_root_t
		size         C.svn_filesize_t
	)

	subPool := initSubPool(r.pool)
	defer C.svn_pool_destroy(subPool)

	if e := C.svn_fs_revision_root(&revisionRoot, r.fs, C.svn_revnum_t(rev), subPool); e != nil {
		return -1, makeError(e)
	}

	cpath := C.CString(path) // convert to C string
	defer C.free(unsafe.Pointer(cpath))

	if e := C.svn_fs_file_length(&size, revisionRoot, cpath, subPool); e != nil {
		return -1, makeError(e)
	}

	sizeC := int64(size)
	// copy data
	size64 := sizeC
	return size64, nil
}

// Returns file mime type
func (r *Repo) MimeType(path string, rev int64) (string, error) {
	var (
		mimetype     *C.svn_string_t
		revisionRoot *C.svn_fs_root_t
	)

	subPool := initSubPool(r.pool)
	defer C.svn_pool_destroy(subPool)

	mime := ""

	if e := C.svn_fs_revision_root(&revisionRoot, r.fs, C.svn_revnum_t(rev), subPool); e != nil {
		return "", makeError(e)
	}

	cpath := C.CString(path) // convert to C string
	defer C.free(unsafe.Pointer(cpath))

	if err := C.FileMimeType(&mimetype, revisionRoot, cpath, subPool); err != nil {
		return "", makeError(err)
	}

	if mimetype != nil {
		mime = C.GoString(mimetype.data)
	}

	return mime, nil
}

func (r *Repo) initSvnStream(fs *C.svn_fs_root_t, pool *C.apr_pool_t, path string) (*SvnStream, error) {
	var (
		stream *C.svn_stream_t
	)

	cpath := C.CString(path) // convert to C string
	defer C.free(unsafe.Pointer(cpath))

	if err := C.svn_fs_file_contents(&stream, fs, cpath, pool); err != nil {
		return nil, makeError(err)
	}

	svnStream := &SvnStream{io: stream, pool: pool}
	runtime.SetFinalizer(svnStream, (*SvnStream).Close)

	return svnStream, nil
}

func initSubPool(pool *C.apr_pool_t) *C.apr_pool_t {
	return C.svn_pool_create_ex(pool, nil)
}
